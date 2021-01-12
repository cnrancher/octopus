package physical

import (
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/JuulLabs-OSS/ble"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/rancher/octopus/adaptors/ble/api/v1alpha1"
	"github.com/rancher/octopus/adaptors/ble/pkg/metadata"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/socket/handler"
	"github.com/rancher/octopus/pkg/mqtt"
	"github.com/rancher/octopus/pkg/util/object"
)

// Device is an interface for device operations set.
type Device interface {
	// Shutdown uses to close the connection between adaptor and real(physical) device.
	Shutdown()
	// Configure uses to set up the device.
	Configure(references api.ReferencesHandler, device *v1alpha1.BluetoothDevice) error
}

// NewDevice creates a Device.
func NewDevice(log logr.Logger, meta metav1.ObjectMeta, toLimb BluetoothDeviceLimSyncer) Device {
	log.Info("Created ")
	return &bleDevice{
		log: log,
		instance: &v1alpha1.BluetoothDevice{
			ObjectMeta: meta,
		},
		toLimb: toLimb,
	}
}

type bleDevice struct {
	sync.Mutex

	log             logr.Logger
	instance        *v1alpha1.BluetoothDevice
	toLimb          BluetoothDeviceLimSyncer
	stop            chan struct{}
	bluetoothClient *BluetoothPeripheral

	mqttClient mqtt.Client
}

func (d *bleDevice) Configure(references api.ReferencesHandler, device *v1alpha1.BluetoothDevice) error {
	defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

	d.Lock()
	defer d.Unlock()

	var newSpec = device.Spec
	var staleSpec = d.instance.Spec

	// configures MQTT client if needed
	var staleExtension, newExtension v1alpha1.BluetoothDeviceExtension
	if d.instance.Spec.Extension != nil {
		staleExtension = *d.instance.Spec.Extension
	}
	if newSpec.Extension != nil {
		newExtension = *newSpec.Extension
	}
	if !reflect.DeepEqual(staleExtension.MQTT, newExtension.MQTT) {
		if d.mqttClient != nil {
			d.mqttClient.Disconnect()
			d.mqttClient = nil
		}

		if newExtension.MQTT != nil {
			var cli, err = mqtt.NewClient(*newExtension.MQTT, object.GetControlledOwnerObjectReference(device), references)
			if err != nil {
				return errors.Wrap(err, "failed to create MQTT client")
			}

			err = cli.Connect()
			if err != nil {
				return errors.Wrap(err, "failed to connect MQTT broker")
			}
			d.mqttClient = cli
		}
	}

	// configures Bluetooth client
	if !reflect.DeepEqual(staleSpec.Protocol, newSpec.Protocol) {
		if d.bluetoothClient != nil {
			if err := d.bluetoothClient.Close(); err != nil {
				if err != io.EOF {
					d.log.Error(err, "Error closing Bluetooth device connection")
				}
			}
			d.bluetoothClient = nil
		}

		var peripherals, err = ScanBluetoothPeripherals(newSpec.Protocol.Endpoint, newSpec.Protocol.GetScanTimeout())
		if err != nil {
			return errors.Wrapf(err, "failed to scan Bluetooth device %s", newSpec.Protocol.Endpoint)
		}

		var autoReconnect = newSpec.Protocol.IsAutoReconnect()
		var peripheral = peripherals[0]
		peripheral.SetConnectionOptions(BluetoothPeripheralConnectionOptions{
			AutoReconnect:                  autoReconnect,
			MaxReconnectInterval:           newSpec.Protocol.GetMaxReconnectInterval(),
			ConnectMTU:                     newSpec.Protocol.GetConnectionMTU(),
			ConnectTimeout:                 newSpec.Protocol.GetConnectTimeout(),
			OnlySubscribeNotificationValue: newSpec.Protocol.OnlySubscribeNotificationValue,
			OnlyWriteValueWithoutResponse:  newSpec.Protocol.OnlyWriteValueWithoutResponse,
			OnConnectionLost: func(_ ble.Client, cerr error) {
				if autoReconnect {
					d.log.Error(errors.Cause(cerr), "Bluetooth device connection is closed, please turn off the AutoReconnect if want to know this at the first time")
					return
				}

				// NB(thxCode) feedbacks the EOF of Bluetooth device connection if turn off the auto reconnection.
				var feedbackErr error
				if cerr != ErrGATTPeripheralConnectionClosed {
					feedbackErr = errors.Wrapf(cerr, "error for Bluetooth device connection")
				} else {
					feedbackErr = errors.New("Bluetooth device connection is closed")
				}
				if d.toLimb != nil {
					if err := d.toLimb(nil, feedbackErr); err != nil {
						d.log.Error(err, "failed to feedback the lost error of Bluetooth device connection")
					}
				}
			},
		})
		d.bluetoothClient = peripheral
		d.bluetoothClient.Start()

		// NB(thxCode) since the client has been changed,
		// we need to reset.
		d.instance.Spec = v1alpha1.BluetoothDeviceSpec{}
	}

	return d.refresh(newSpec)
}

func (d *bleDevice) Shutdown() {
	d.Lock()
	defer d.Unlock()

	d.stopFetch()
	if d.bluetoothClient != nil {
		if err := d.bluetoothClient.Close(); err != nil {
			if errors.Cause(err) != io.EOF {
				d.log.Error(err, "Error closing Bluetooth connection")
			}
		}
		d.bluetoothClient = nil
	}
	if d.mqttClient != nil {
		d.mqttClient.Disconnect()
		d.mqttClient = nil
	}
	d.log.Info("Shutdown")
}

// refresh refreshes the status with new spec.
func (d *bleDevice) refresh(newSpec v1alpha1.BluetoothDeviceSpec) error {
	var newStatus v1alpha1.BluetoothDeviceStatus

	var staleStatus = d.instance.Status
	var staleSpec = d.instance.Spec
	if !reflect.DeepEqual(staleSpec.Properties, newSpec.Properties) {
		d.stopFetch()

		var staleSpecPropsMap = mapSpecProperties(staleSpec.Properties)
		var staleStatusPropsMap = mapStatusProperties(staleStatus.Properties)

		// syncs properties
		var statusProps = make([]v1alpha1.BluetoothDeviceStatusProperty, 0, len(newSpec.Properties))
		for i := 0; i < len(newSpec.Properties); i++ {
			var specPropPtr = &newSpec.Properties[i]
			var statusProp v1alpha1.BluetoothDeviceStatusProperty
			if staleStatusPropPtr, existed := staleStatusPropsMap[specPropPtr.Name]; existed {
				statusProp = *staleStatusPropPtr
			}

			for _, accessMode := range specPropPtr.MergeAccessModes() {
				switch accessMode {
				case v1alpha1.BluetoothDevicePropertyAccessModeNotify:
					var statusPropPtr, err = d.subscribeProperty(specPropPtr, i)
					if err != nil {
						return errors.Wrapf(err, "failed to notify property %s", specPropPtr.Name)
					}
					statusProp = *statusPropPtr
				case v1alpha1.BluetoothDevicePropertyAccessModeWriteOnce:
					if !reflect.DeepEqual(specPropPtr, staleSpecPropsMap[specPropPtr.Name]) {
						var statusPropPtr, err = d.writeProperty(specPropPtr, statusProp.UpdatedAt)
						if err != nil {
							return errors.Wrapf(err, "failed to write property %s", specPropPtr.Name)
						}
						statusProp = *statusPropPtr
					}
				case v1alpha1.BluetoothDevicePropertyAccessModeWriteMany:
					var statusPropPtr, err = d.writeProperty(specPropPtr, statusProp.UpdatedAt)
					if err != nil {
						return errors.Wrapf(err, "failed to write property %s", specPropPtr.Name)
					}
					statusProp = *statusPropPtr
				case v1alpha1.BluetoothDevicePropertyAccessModeReadOnce:
					if !reflect.DeepEqual(specPropPtr, staleSpecPropsMap[specPropPtr.Name]) {
						var statusPropPtr, err = d.readProperty(specPropPtr)
						if err != nil {
							return errors.Wrapf(err, "failed to read property %s", specPropPtr.Name)
						}
						statusProp = *statusPropPtr
					}
				default: // BluetoothDevicePropertyAccessModeReadMany
					var statusPropPtr, err = d.readProperty(specPropPtr)
					if err != nil {
						return errors.Wrapf(err, "failed to read property %s", specPropPtr.Name)
					}
					statusProp = *statusPropPtr
				}
			}

			statusProps = append(statusProps, statusProp)
		}
		newStatus = v1alpha1.BluetoothDeviceStatus{Properties: statusProps}
	} else {
		newStatus = staleStatus
	}

	// fetches in backend
	d.startFetch(newSpec.Protocol.GetSyncInterval())

	// records
	d.instance.Spec = newSpec
	d.instance.Status = newStatus
	return d.sync()
}

// writeProperty writes data of a property to device.
func (d *bleDevice) writeProperty(propPtr *v1alpha1.BluetoothDeviceProperty, updatedAt *metav1.Time) (*v1alpha1.BluetoothDeviceStatusProperty, error) {
	if propPtr.Value != "" {
		var data, err = convertValueToBytes(propPtr)
		if err != nil {
			return nil, err
		}

		err = d.bluetoothClient.WriteCharacteristic(propPtr.Visitor.Service, propPtr.Visitor.Characteristic, data)
		if err != nil {
			return nil, errors.Wrap(err, "failed to write")
		}
		d.log.V(4).Info("Write property", "property", propPtr.Name, "type", propPtr.Type, "value", propPtr.Value)
		updatedAt = now() // updates the timestamp
	}

	if updatedAt == nil {
		updatedAt = now() // records current timestamp
	}
	var statusPropPtr = &v1alpha1.BluetoothDeviceStatusProperty{
		Name:        propPtr.Name,
		Type:        propPtr.Type,
		AccessModes: propPtr.AccessModes,
		UpdatedAt:   updatedAt,
	}
	return statusPropPtr, nil
}

// readProperty reads data of a property from device.
func (d *bleDevice) readProperty(propPtr *v1alpha1.BluetoothDeviceProperty) (*v1alpha1.BluetoothDeviceStatusProperty, error) {
	var data, err = d.bluetoothClient.ReadCharacteristic(propPtr.Visitor.Service, propPtr.Visitor.Characteristic)
	if err != nil {
		return nil, err
	}
	value, operationResult, err := parseValueFromBytes(data, propPtr)
	if err != nil {
		return nil, err
	}
	d.log.V(4).Info("Read property", "property", propPtr.Name, "type", propPtr.Type, "value", value, "operationResult", operationResult)

	var statusPropPtr = &v1alpha1.BluetoothDeviceStatusProperty{
		Name:            propPtr.Name,
		Type:            propPtr.Type,
		AccessModes:     propPtr.AccessModes,
		Value:           value,
		OperationResult: operationResult,
		UpdatedAt:       now(),
	}
	return statusPropPtr, nil
}

// subscribeProperty subscribes a property to receive the changes from device.
func (d *bleDevice) subscribeProperty(propPtr *v1alpha1.BluetoothDeviceProperty, index int) (*v1alpha1.BluetoothDeviceStatusProperty, error) {
	var receiver = func(data []byte) {
		d.Lock()
		defer d.Unlock()

		if index >= len(d.instance.Status.Properties) {
			return
		}

		var value, operationResult, err = parseValueFromBytes(data, propPtr)
		if err != nil {
			// TODO give a way to feedback this to limb.
			d.log.Error(err, "Error converting the byte array to property value")
			return
		}
		d.log.V(4).Info("Notify property", "property", propPtr.Name, "type", propPtr.Type, "value", value, "operationResult", operationResult)

		var statusPropPtr = &v1alpha1.BluetoothDeviceStatusProperty{
			Name:            propPtr.Name,
			Type:            propPtr.Type,
			AccessModes:     propPtr.AccessModes,
			Value:           value,
			OperationResult: operationResult,
			UpdatedAt:       now(),
		}
		d.instance.Status.Properties[index] = *statusPropPtr

		// TODO we need to debounce here
		if err := d.sync(); err != nil {
			d.log.Error(err, "Failed to sync")
		}
	}

	var err = d.bluetoothClient.SubscribeCharacteristic(
		propPtr.Visitor.Service,
		propPtr.Visitor.Characteristic,
		receiver,
	)
	if err != nil {
		return nil, err
	}

	var statusPropPtr = &v1alpha1.BluetoothDeviceStatusProperty{
		Name:        propPtr.Name,
		Type:        propPtr.Type,
		AccessModes: propPtr.AccessModes,
		UpdatedAt:   now(),
	}
	return statusPropPtr, nil
}

// fetch is blocked, it is used to sync the Bluetooth device status periodically,
// it's worth noting that it just reads or writes the "ReadMany/WriteMany" properties.
func (d *bleDevice) fetch(interval time.Duration, stop <-chan struct{}) {
	defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

	d.log.Info("Fetching")
	defer func() {
		d.log.Info("Finished fetching")
	}()

	var ticker = time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
		}

		d.Lock()
		func() {
			defer d.Unlock()

			// NB(thxCode) when the `spec.protocol` changes,
			// the `spec.properties` will be reset,
			// after obtaining the lock, this `fetch` goroutine should end.
			if len(d.instance.Status.Properties) != len(d.instance.Spec.Properties) {
				return
			}

			for i, statusProp := range d.instance.Status.Properties {
				var specPropPtr = &d.instance.Spec.Properties[i]

				for _, accessMode := range specPropPtr.MergeAccessModes() {
					switch accessMode {
					case v1alpha1.BluetoothDevicePropertyAccessModeWriteMany:
						var statusPropPtr, err = d.writeProperty(specPropPtr, statusProp.UpdatedAt)
						if err != nil {
							// TODO give a way to feedback this to limb.
							d.log.Error(err, "Error writing property", "property", statusProp.Name)
							continue
						}
						statusProp = *statusPropPtr
					case v1alpha1.BluetoothDevicePropertyAccessModeReadMany:
						var statusPropPtr, err = d.readProperty(specPropPtr)
						if err != nil {
							// TODO give a way to feedback this to limb.
							d.log.Error(err, "Error reading property", "property", statusProp.Name)
							continue
						}
						statusProp = *statusPropPtr
					default:
						continue
					}
				}

				d.instance.Status.Properties[i] = statusProp
			}
			if err := d.sync(); err != nil {
				d.log.Error(err, "Failed to sync")
			}
		}()

		select {
		case <-d.stop:
			return
		default:
		}
	}
}

// stopFetch stops the asynchronous fetch.
func (d *bleDevice) stopFetch() {
	if d.stop != nil {
		close(d.stop)
		d.stop = nil
	}

	// unsubscribe all characteristics
	if d.bluetoothClient != nil {
		if err := d.bluetoothClient.ClearSubscriptions(); err != nil && err != ErrGATTPeripheralConnectionClosed {
			d.log.Error(err, "Failed to unsubscribe all characteristics")
		}
	}
}

// startFetch starts the asynchronous fetch.
func (d *bleDevice) startFetch(interval time.Duration) {
	if d.stop == nil {
		d.stop = make(chan struct{})
		go d.fetch(interval, d.stop)
	}
}

// sync combines all synchronization operations.
func (d *bleDevice) sync() error {
	if d.toLimb != nil {
		if err := d.toLimb(d.instance, nil); err != nil {
			return err
		}
	}
	if d.mqttClient != nil {
		if err := d.mqttClient.Publish(mqtt.PublishMessage{Payload: d.instance.Status}); err != nil {
			return err
		}
	}
	d.log.V(1).Info("Synced")
	return nil
}

func mapSpecProperties(specProps []v1alpha1.BluetoothDeviceProperty) map[string]*v1alpha1.BluetoothDeviceProperty {
	var ret = make(map[string]*v1alpha1.BluetoothDeviceProperty, len(specProps))
	for i := 0; i < len(specProps); i++ {
		var prop = specProps[i]
		ret[prop.Name] = &prop
	}
	return ret
}

func mapStatusProperties(statusProps []v1alpha1.BluetoothDeviceStatusProperty) map[string]*v1alpha1.BluetoothDeviceStatusProperty {
	var ret = make(map[string]*v1alpha1.BluetoothDeviceStatusProperty, len(statusProps))
	for i := 0; i < len(statusProps); i++ {
		var prop = statusProps[i]
		ret[prop.Name] = &prop
	}
	return ret
}

func now() *metav1.Time {
	var ret = metav1.Now()
	return &ret
}

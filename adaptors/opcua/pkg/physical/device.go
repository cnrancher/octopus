package physical

import (
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/gopcua/opcua/ua"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/rancher/octopus/adaptors/opcua/api/v1alpha1"
	"github.com/rancher/octopus/adaptors/opcua/pkg/metadata"
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
	Configure(references api.ReferencesHandler, obj *v1alpha1.OPCUADevice) error
}

// NewDevice creates a Device.
func NewDevice(log logr.Logger, meta metav1.ObjectMeta, toLimb OPCUADeviceLimbSyncer) Device {
	log.Info("Created ")
	return &opcuaDevice{
		log: log,
		instance: &v1alpha1.OPCUADevice{
			ObjectMeta: meta,
		},
		toLimb: toLimb,
	}
}

type opcuaDevice struct {
	sync.Mutex

	log         logr.Logger
	instance    *v1alpha1.OPCUADevice
	toLimb      OPCUADeviceLimbSyncer
	stop        chan struct{}
	opcuaClient *OPCUAClient

	mqttClient mqtt.Client
}

func (d *opcuaDevice) Configure(references api.ReferencesHandler, device *v1alpha1.OPCUADevice) error {
	defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

	d.Lock()
	defer d.Unlock()

	var newSpec = device.Spec
	var staleSpec = d.instance.Spec

	// configures MQTT opcuaClient if needed
	var staleExtension, newExtension v1alpha1.OPCUADeviceExtension
	if staleSpec.Extension != nil {
		staleExtension = *staleSpec.Extension
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
				return errors.Wrap(err, "failed to create MQTT opcuaClient")
			}

			err = cli.Connect()
			if err != nil {
				return errors.Wrap(err, "failed to connect MQTT broker")
			}
			d.mqttClient = cli
		}
	}

	// configures OPC-UA client
	if !reflect.DeepEqual(staleSpec.Protocol, newSpec.Protocol) {
		if d.opcuaClient != nil {
			if err := d.opcuaClient.Close(); err != nil {
				if err != io.EOF {
					d.log.Error(err, "Error closing OPC-UA connection")
				}
			}
			d.opcuaClient = nil
		}

		var client, err = NewOPCUAClient(newSpec.Protocol, references)
		if err != nil {
			return errors.Wrap(err, "failed to create OPC-UA client")
		}
		d.opcuaClient = client

		// NB(thxCode) since the client has been changed,
		// we need to reset.
		d.instance.Spec = v1alpha1.OPCUADeviceSpec{}
	}

	return d.refresh(newSpec)
}

func (d *opcuaDevice) Shutdown() {
	d.Lock()
	defer d.Unlock()

	d.stopFetch()
	if d.opcuaClient != nil {
		if err := d.opcuaClient.Close(); err != nil {
			if err != io.EOF {
				d.log.Error(err, "Error closing OPC-UA connection")
			}
		}
		d.opcuaClient = nil
	}
	if d.mqttClient != nil {
		d.mqttClient.Disconnect()
		d.mqttClient = nil
	}
	d.log.Info("Shutdown")
}

// refresh refreshes the status with new spec.
func (d *opcuaDevice) refresh(newSpec v1alpha1.OPCUADeviceSpec) error {
	var newStatus v1alpha1.OPCUADeviceStatus

	var staleStatus = d.instance.Status
	var staleSpec = d.instance.Spec
	if !reflect.DeepEqual(staleSpec.Properties, newSpec.Properties) {
		d.stopFetch()

		var staleSpecPropsMap = mapSpecProperties(staleSpec.Properties)
		var staleStatusPropsMap = mapStatusProperties(staleStatus.Properties)

		// syncs properties
		var statusProps = make([]v1alpha1.OPCUADeviceStatusProperty, 0, len(newSpec.Properties))
		for i := 0; i < len(newSpec.Properties); i++ {
			var specPropPtr = &newSpec.Properties[i]
			var statusProp v1alpha1.OPCUADeviceStatusProperty
			if staleStatusPropPtr, existed := staleStatusPropsMap[specPropPtr.Name]; existed {
				statusProp = *staleStatusPropPtr
			}

			for _, accessMode := range specPropPtr.MergeAccessModes() {
				switch accessMode {
				case v1alpha1.OPCUADevicePropertyAccessModeNotify:
					var statusPropPtr, err = d.subscribeProperty(specPropPtr, i)
					if err != nil {
						return errors.Wrapf(err, "failed to notify property %s", specPropPtr.Name)
					}
					statusProp = *statusPropPtr
				case v1alpha1.OPCUADevicePropertyAccessModeWriteOnce:
					if !reflect.DeepEqual(specPropPtr, staleSpecPropsMap[specPropPtr.Name]) {
						var statusPropPtr, err = d.writeProperty(specPropPtr, statusProp.UpdatedAt)
						if err != nil {
							return errors.Wrapf(err, "failed to write property %s", specPropPtr.Name)
						}
						statusProp = *statusPropPtr
					}
				case v1alpha1.OPCUADevicePropertyAccessModeWriteMany:
					var statusPropPtr, err = d.writeProperty(specPropPtr, statusProp.UpdatedAt)
					if err != nil {
						return errors.Wrapf(err, "failed to write property %s", specPropPtr.Name)
					}
					statusProp = *statusPropPtr
				case v1alpha1.OPCUADevicePropertyAccessModeReadOnce:
					if !reflect.DeepEqual(specPropPtr, staleSpecPropsMap[specPropPtr.Name]) {
						var statusPropPtr, err = d.readProperty(specPropPtr)
						if err != nil {
							return errors.Wrapf(err, "failed to read property %s", specPropPtr.Name)
						}
						statusProp = *statusPropPtr
					}
				default: // OPCUADevicePropertyAccessModeReadMany
					var statusPropPtr, err = d.readProperty(specPropPtr)
					if err != nil {
						return errors.Wrapf(err, "failed to read property %s", specPropPtr.Name)
					}
					statusProp = *statusPropPtr
				}
			}

			statusProps = append(statusProps, statusProp)
		}
		newStatus = v1alpha1.OPCUADeviceStatus{Properties: statusProps}
	} else {
		newStatus = staleStatus
	}

	// fetches in backend
	if err := d.startFetch(newSpec.Protocol.GetSyncInterval()); err != nil {
		return errors.Wrap(err, "failed to start fetch")
	}

	// records
	d.instance.Spec = newSpec
	d.instance.Status = newStatus
	return d.sync()
}

// writeProperty writes data of a property to device.
func (d *opcuaDevice) writeProperty(propPtr *v1alpha1.OPCUADeviceProperty, updatedAt *metav1.Time) (*v1alpha1.OPCUADeviceStatusProperty, error) {
	if propPtr.Value != "" {
		var value, err = convertValueToVariant(propPtr)
		if err != nil {
			return nil, err
		}

		err = d.opcuaClient.WriteDataValue(
			propPtr.Visitor.NodeID,
			&ua.DataValue{
				EncodingMask: ua.DataValueValue,
				Value:        value,
			},
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to write")
		}
		d.log.V(4).Info("Write property", "property", propPtr.Name, "type", propPtr.Type, "value", propPtr.Value)
		updatedAt = now() // updates the timestamp
	}

	if updatedAt == nil {
		updatedAt = now() // records current timestamp
	}
	var statusPropPtr = &v1alpha1.OPCUADeviceStatusProperty{
		Name:        propPtr.Name,
		Type:        propPtr.Type,
		AccessModes: propPtr.AccessModes,
		UpdatedAt:   updatedAt,
	}
	return statusPropPtr, nil
}

// readProperty reads data of a property from device.
func (d *opcuaDevice) readProperty(propPtr *v1alpha1.OPCUADeviceProperty) (*v1alpha1.OPCUADeviceStatusProperty, error) {
	var data, err = d.opcuaClient.ReadDataValue(propPtr.Visitor.NodeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read")
	}

	value, operationResult, err := parseValueFromVariant(data.Value, propPtr)
	if err != nil {
		return nil, err
	}
	d.log.V(4).Info("Read property", "property", propPtr.Name, "type", propPtr.Type, "value", value, "operationResult", operationResult)

	var statusPropPtr = &v1alpha1.OPCUADeviceStatusProperty{
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
func (d *opcuaDevice) subscribeProperty(propPtr *v1alpha1.OPCUADeviceProperty, index int) (*v1alpha1.OPCUADeviceStatusProperty, error) {
	var receiver = func(data *ua.DataValue) {
		d.Lock()
		defer d.Unlock()

		if index >= len(d.instance.Status.Properties) {
			return
		}

		var value, operationResult, err = parseValueFromVariant(data.Value, propPtr)
		if err != nil {
			// TODO give a way to feedback this to limb.
			d.log.Error(err, "Error converting the byte array to property value")
			return
		}
		d.log.V(4).Info("Notify property", "property", propPtr.Name, "type", propPtr.Type, "value", value, "operationResult", operationResult)

		var statusPropPtr = &v1alpha1.OPCUADeviceStatusProperty{
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

	var err = d.opcuaClient.RegisterDataValueSubscription(
		propPtr.Visitor.NodeID,
		receiver,
	)
	if err != nil {
		return nil, err
	}

	var statusPropPtr = &v1alpha1.OPCUADeviceStatusProperty{
		Name:        propPtr.Name,
		Type:        propPtr.Type,
		AccessModes: propPtr.AccessModes,
		UpdatedAt:   now(),
	}
	return statusPropPtr, nil
}

// fetch is blocked, it is used to sync the OPU-UA device status periodically,
// it's worth noting that it just reads or writes the "ReadMany/WriteMany" properties.
func (d *opcuaDevice) fetch(interval time.Duration, stop <-chan struct{}) {
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
					case v1alpha1.OPCUADevicePropertyAccessModeWriteMany:
						var statusPropPtr, err = d.writeProperty(specPropPtr, statusProp.UpdatedAt)
						if err != nil {
							// TODO give a way to feedback this to limb.
							d.log.Error(err, "Error writing property", "property", statusProp.Name)
							continue
						}
						statusProp = *statusPropPtr
					case v1alpha1.OPCUADevicePropertyAccessModeReadMany:
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
func (d *opcuaDevice) stopFetch() {
	if d.stop == nil {
		return
	}

	// closes fetching
	close(d.stop)
	d.stop = nil

	// stops all UA subscriptions
	if d.opcuaClient != nil {
		d.opcuaClient.StopSubscriptions()
	}
}

// startFetch starts the asynchronous fetch.
func (d *opcuaDevice) startFetch(interval time.Duration) error {
	if d.stop != nil {
		return nil
	}
	d.stop = make(chan struct{})

	// starts fetching
	go d.fetch(interval, d.stop)

	// starts all UA subscriptions
	if d.opcuaClient != nil {
		if err := d.opcuaClient.StartSubscriptions(d.stop); err != nil {
			return err
		}
	}

	return nil
}

// sync combines all synchronization operations.
func (d *opcuaDevice) sync() error {
	if d.toLimb != nil {
		if err := d.toLimb(d.instance); err != nil {
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

func mapSpecProperties(specProps []v1alpha1.OPCUADeviceProperty) map[string]*v1alpha1.OPCUADeviceProperty {
	var ret = make(map[string]*v1alpha1.OPCUADeviceProperty, len(specProps))
	for i := 0; i < len(specProps); i++ {
		var prop = specProps[i]
		ret[prop.Name] = &prop
	}
	return ret
}

func mapStatusProperties(statusProps []v1alpha1.OPCUADeviceStatusProperty) map[string]*v1alpha1.OPCUADeviceStatusProperty {
	var ret = make(map[string]*v1alpha1.OPCUADeviceStatusProperty, len(statusProps))
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

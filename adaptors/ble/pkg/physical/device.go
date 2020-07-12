package physical

import (
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rancher/octopus/pkg/mqtt"
	"github.com/rancher/octopus/pkg/util/object"

	"github.com/bettercap/gatt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"

	"github.com/rancher/octopus/adaptors/ble/api/v1alpha1"
)

// Device is an interface for device operations set.
type Device interface {
	// Shutdown uses to close the connection between adaptor and real(physical) device.
	Shutdown()
	// Configure uses to set up the device.
	Configure(references api.ReferencesHandler, device *v1alpha1.BluetoothDevice) error
}

// NewDevice creates a Device.
func NewDevice(log logr.Logger, meta metav1.ObjectMeta, toLimb BluetoothDeviceLimSyncer, gattDevice gatt.Device) Device {
	log.Info("Created ")
	return &bleDevice{
		log: log,
		instance: &v1alpha1.BluetoothDevice{
			ObjectMeta: meta,
		},
		toLimb:     toLimb,
		gattDevice: gattDevice,
	}
}

type bleDevice struct {
	sync.Mutex

	stop chan struct{}
	wg   sync.WaitGroup

	log        logr.Logger
	instance   *v1alpha1.BluetoothDevice
	name       types.NamespacedName
	toLimb     BluetoothDeviceLimSyncer
	gattDevice gatt.Device

	mqttClient mqtt.Client
}

func (d *bleDevice) Configure(references api.ReferencesHandler, device *v1alpha1.BluetoothDevice) error {
	d.Lock()
	defer d.Unlock()

	var newSpec = device.Spec

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

	return d.refresh(newSpec)
}

func (d *bleDevice) Shutdown() {
	d.Lock()
	defer d.Unlock()

	d.stopFetch()
	if d.mqttClient != nil {
		d.mqttClient.Disconnect()
		d.mqttClient = nil
	}
	d.log.Info("Shutdown")
}

// refresh refreshes the status with new spec.
func (d *bleDevice) refresh(newSpec v1alpha1.BluetoothDeviceSpec) error {
	var status = d.instance.Status
	var staleSpec = d.instance.Spec
	if !reflect.DeepEqual(staleSpec.Protocol, newSpec.Protocol) ||
		!reflect.DeepEqual(staleSpec.Parameters, newSpec.Parameters) {
		d.stopFetch()

		var statusProps, err = d.scanDevice(newSpec)
		if err != nil {
			return err
		}
		status = v1alpha1.BluetoothDeviceStatus{Properties: statusProps}
	}

	// fetches in backend
	d.startFetch(newSpec.Parameters.SyncInterval.Duration)

	// records
	d.instance.Spec = newSpec
	d.instance.Status = status
	return d.sync()
}

// fetch is blocked, it is used to sync the ble device status periodically,
// it's worth noting that it just reads the properties from bel device.
func (d *bleDevice) fetch(interval time.Duration, stop <-chan struct{}) {
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

			var props, err = d.scanDevice(d.instance.Spec)
			if err != nil {
				// TODO give a way to feedback this to limb.
				d.log.Error(err, "failed to scan device")
			} else {
				d.instance.Status.Properties = props
			}
			if err := d.sync(); err != nil {
				d.log.Error(err, "failed to sync")
			}
		}()

		select {
		case <-d.stop:
			return
		default:
		}
	}
}

func (d *bleDevice) scanDevice(spec v1alpha1.BluetoothDeviceSpec) ([]v1alpha1.BluetoothDeviceStatusProperty, error) {
	d.log.V(4).Info("Scanning device")

	var ctrl = &BLEController{
		endpoint:   spec.Protocol.Endpoint,
		properties: spec.Properties,
		done:       make(chan struct{}),
		log:        d.log,
	}
	// register BLE device handlers.
	d.gattDevice.Handle(
		gatt.PeripheralDiscovered(ctrl.onPeripheralDiscovered),
		gatt.PeripheralConnected(ctrl.onPeripheralConnected),
		gatt.PeripheralDisconnected(ctrl.onPeriphDisconnected),
	)
	d.gattDevice.Init(ctrl.onStateChanged)

	var timeout = time.NewTimer(spec.Parameters.Timeout.Duration)
	defer timeout.Stop()
	select {
	case <-timeout.C:
		return nil, errors.Errorf("timeout to scan device in %s", spec.Parameters.Timeout.Duration)
	case <-ctrl.done:
	}

	d.log.V(4).Info("Finished scanning")
	return ctrl.statusProps, nil
}

func (d *bleDevice) stopFetch() {
	if d.stop != nil {
		close(d.stop)
		d.stop = nil
	}
}

func (d *bleDevice) startFetch(interval time.Duration) {
	if d.stop == nil {
		d.stop = make(chan struct{})
		go d.fetch(interval, d.stop)
	}
}

// sync combines all synchronization operations.
func (d *bleDevice) sync() error {
	if err := d.toLimb(d.instance); err != nil {
		return err
	}
	if d.mqttClient != nil {
		if err := d.mqttClient.Publish(mqtt.PublishMessage{Payload: d.instance.Status}); err != nil {
			return err
		}
	}
	d.log.V(1).Info("Synced")
	return nil
}

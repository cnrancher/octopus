package physical

import (
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/rancher/octopus/pkg/mqtt"
	"github.com/rancher/octopus/pkg/util/object"

	"github.com/bettercap/gatt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"

	"github.com/rancher/octopus/adaptors/ble/api/v1alpha1"
)

type Device interface {
	Configure(references api.ReferencesHandler, obj v1alpha1.BluetoothDevice, gattDevice gatt.Device) error
	Shutdown()
}

type device struct {
	sync.Mutex

	stop chan struct{}
	wg   sync.WaitGroup

	log     logr.Logger
	name    types.NamespacedName
	handler DataHandler

	spec   v1alpha1.BluetoothDeviceSpec
	status v1alpha1.BluetoothDeviceStatus

	mqttClient mqtt.Client
}

func NewDevice(log logr.Logger, name types.NamespacedName, handler DataHandler) Device {
	return &device{
		log:     log,
		name:    name,
		handler: handler,
	}
}

func (d *device) Configure(references api.ReferencesHandler, obj v1alpha1.BluetoothDevice, gatt gatt.Device) error {
	spec := d.spec
	d.spec = obj.Spec

	// configure protocol config and parameters
	if !reflect.DeepEqual(d.spec.Protocol, spec.Protocol) || !reflect.DeepEqual(d.spec.Parameters, spec.Parameters) {
		d.connect(gatt)
	}

	if !reflect.DeepEqual(d.spec.Extension, spec.Extension) {
		if d.mqttClient != nil {
			d.mqttClient.Disconnect()
			d.mqttClient = nil
		}

		if d.spec.Extension.MQTT != nil {
			var cli, err = mqtt.NewClient(*d.spec.Extension.MQTT, object.GetControlledOwnerObjectReference(&obj), references)
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
	return nil
}

func (d *device) connect(gattDevice gatt.Device) {
	if d.stop != nil {
		close(d.stop)
	}
	d.stop = make(chan struct{})

	d.log.Info("Connect to device", "device", d.name)

	// run periodically to sync device status
	var ticker = time.NewTicker(d.spec.Parameters.SyncInterval.Duration)
	defer ticker.Stop()
	d.log.V(2).Info("Sync interval is set to", d.spec.Parameters.SyncInterval)

	go func() {
		for {
			cont := BLEController{
				spec:   d.spec,
				status: d.status,
				done:   make(chan struct{}),
				log:    d.log,
			}
			// Register BLE device handlers.
			gattDevice.Handle(
				gatt.PeripheralDiscovered(cont.onPeripheralDiscovered),
				gatt.PeripheralConnected(cont.onPeripheralConnected),
				gatt.PeripheralDisconnected(cont.onPeriphDisconnected),
			)

			gattDevice.Init(cont.onStateChanged)
			<-cont.done
			d.log.Info("Device Done")

			d.handler(d.name, cont.status)
			d.log.Info("Synced ble device status", cont.status)

			// pub updated status to the MQTT broker
			if d.mqttClient != nil {
				var status = cont.status.DeepCopy()
				if err := d.mqttClient.Publish(mqtt.PublishMessage{Payload: status}); err != nil {
					d.log.Error(err, "Failed to publish MQTT message")
				}
			}
			d.log.V(2).Info("Success pub device status to the MQTT Broker", cont.status.Properties)

			select {
			case <-d.stop:
				return
			default:
			}
		}
	}()
}

func (d *device) Shutdown() {
	d.Lock()
	defer d.Unlock()

	if d.stop != nil {
		close(d.stop)
	}

	// close MQTT connection
	if d.mqttClient != nil {
		d.mqttClient.Disconnect()
		d.mqttClient = nil
	}

	d.log.Info("Closed connection")
}

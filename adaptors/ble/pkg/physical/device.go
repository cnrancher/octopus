package physical

import (
	"reflect"
	"sync"
	"time"

	"github.com/bettercap/gatt"
	"github.com/go-logr/logr"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"

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

			select {
			case <-d.stop:
				return
			default:
			}
		}
	}()
}

func (d *device) Shutdown() {
	if d.stop != nil {
		close(d.stop)
	}
	d.log.Info("Closed connection")
}

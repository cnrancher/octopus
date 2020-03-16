package physical

import (
	"sync"
	"time"

	"github.com/bettercap/gatt"
	"github.com/go-logr/logr"
	"github.com/rancher/octopus/adaptors/ble/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
)

type Device interface {
	Configure(spec v1alpha1.BluetoothDeviceSpec, status v1alpha1.BluetoothDeviceStatus)
	Shutdown()
}
type device struct {
	sync.Mutex

	stop chan struct{}

	log     logr.Logger
	name    types.NamespacedName
	handler DataHandler

	syncInterval time.Duration
	timeout      time.Duration
	gattDevice   gatt.Device
	properties   []v1alpha1.DeviceProperty
	status       v1alpha1.BluetoothDeviceStatus
}

func NewDevice(log logr.Logger, name types.NamespacedName, handler DataHandler, param Parameters,
	gattDevice gatt.Device) Device {
	return &device{
		log:          log,
		name:         name,
		handler:      handler,
		syncInterval: param.syncInterval,
		timeout:      param.timeout,
		gattDevice:   gattDevice,
	}
}

func (d *device) Configure(spec v1alpha1.BluetoothDeviceSpec, status v1alpha1.BluetoothDeviceStatus) {
	logrus.Infof("trying to connect to device: %s\n", spec.Name)
	d.connect(spec, status)
}

func (d *device) connect(spec v1alpha1.BluetoothDeviceSpec, status v1alpha1.BluetoothDeviceStatus) {
	if d.stop != nil {
		close(d.stop)
	}
	d.stop = make(chan struct{})

	var ticker = time.NewTicker(d.syncInterval * time.Second)
	defer ticker.Stop()
	logrus.Infof("sync interval is set to %s", d.syncInterval.String())

	// run periodically to sync device status
	for {
		cont := Controller{
			spec:   spec,
			status: status,
			done:   make(chan struct{}),
		}
		// Register BLE device handlers.
		go d.gattDevice.Handle(
			gatt.PeripheralDiscovered(cont.onPeripheralDiscovered),
			gatt.PeripheralConnected(cont.onPeripheralConnected),
			gatt.PeripheralDisconnected(cont.onPeriphDisconnected),
		)

		d.gattDevice.Init(onStateChanged)
		logrus.Info("Device Done")

		d.handler(d.name, cont.status)
		logrus.Infof("Synced ble device status: %+v", cont.status)
		<-cont.done

		select {
		case <-d.stop:
			return
		case <-ticker.C:
		}
	}
}

func (d *device) Shutdown() {
	if d.stop != nil {
		close(d.stop)
	}
	d.log.Info("closed connection")
}

package physical

import (
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rancher/octopus/adaptors/dummy/api/v1alpha1"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/mqtt"
)

func NewSpecialDevice(log logr.Logger, instance *v1alpha1.DummySpecialDevice, toLimb SpecialDeviceSyncer) Device {
	return &specialDevice{
		log: log,
		instance: &v1alpha1.DummySpecialDevice{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: instance.Namespace,
				Name:      instance.Name,
				UID:       instance.UID,
			},
		},
		toLimb: toLimb,
	}
}

type specialDevice struct {
	sync.Mutex

	stop chan struct{}
	log  logr.Logger

	instance   *v1alpha1.DummySpecialDevice
	toLimb     SpecialDeviceSyncer
	mqttClient mqtt.Client
}

func (d *specialDevice) Configure(references api.ReferencesHandler, configuration interface{}) error {
	var spec, ok = configuration.(v1alpha1.DummySpecialDeviceSpec)
	if !ok {
		d.log.Error(errors.New("invalidate configuration type"), "Failed to configure")
		return nil
	}

	if spec.Gear == "" {
		spec.Gear = v1alpha1.DummySpecialDeviceGearSlow
	}

	d.Lock()
	defer d.Unlock()

	if !reflect.DeepEqual(d.instance.Spec.Extension.MQTT, spec.Extension.MQTT) {
		if d.mqttClient != nil {
			d.mqttClient.Disconnect(5 * time.Second)
			d.mqttClient = nil

			// since there is only a MQTT inside extension field, here can set to nil directly.
			d.instance.Status.Extension = nil
		}

		if spec.Extension.MQTT != nil {
			var cli, outline, err = mqtt.NewClient(d.instance, *spec.Extension.MQTT, references.ToDataMap())
			if err != nil {
				return errors.Wrap(err, "failed to create MQTT client")
			}

			err = cli.Connect()
			if err != nil {
				return errors.Wrap(err, "failed to connect MQTT broker")
			}
			d.mqttClient = cli

			if d.instance.Status.Extension == nil {
				d.instance.Status.Extension = &v1alpha1.DeviceExtensionStatus{}
			}
			d.instance.Status.Extension.MQTT = outline
		}
	}

	d.instance.Spec = spec
	if spec.On {
		d.on(spec.Gear)
	} else {
		d.off()
	}
	return nil
}

func (d *specialDevice) Shutdown() {
	d.Lock()
	defer d.Unlock()

	if d.mqttClient != nil {
		d.mqttClient.Disconnect(5 * time.Second)
		d.mqttClient = nil
	}

	d.off()
	d.log.Info("Closed connection")
}

func (d *specialDevice) on(gear v1alpha1.DummySpecialDeviceGear) {
	if d.instance.Status.Gear == gear {
		return
	}

	if d.stop != nil {
		close(d.stop)
		d.stop = nil
	}
	d.stop = make(chan struct{})

	// setup
	d.instance.Status.Gear = gear
	switch gear {
	case v1alpha1.DummySpecialDeviceGearFast:
		d.instance.Status.RotatingSpeed = 200
	case v1alpha1.DummySpecialDeviceGearMiddle:
		d.instance.Status.RotatingSpeed = 100
	case v1alpha1.DummySpecialDeviceGearSlow:
		d.instance.Status.RotatingSpeed = 0
	}
	d.sync()

	go d.mockPhysicalWatching(gear, d.stop)
}

func (d *specialDevice) off() {
	if d.stop != nil {
		close(d.stop)
		d.stop = nil
	}

	d.instance.Status = v1alpha1.DummySpecialDeviceStatus{}
	d.sync()
}

// mockPhysicalWatching is used to simulate real device state changes
// and synchronize the changed values back to the limb.
func (d *specialDevice) mockPhysicalWatching(gear v1alpha1.DummySpecialDeviceGear, stop <-chan struct{}) {
	d.log.Info("Mocking started")
	defer func() {
		d.log.Info("Mocking finished")
	}()

	var duration time.Duration
	switch gear {
	case v1alpha1.DummySpecialDeviceGearSlow:
		duration = 3 * time.Second
	case v1alpha1.DummySpecialDeviceGearMiddle:
		duration = 2 * time.Second
	case v1alpha1.DummySpecialDeviceGearFast:
		duration = 1 * time.Second
	}
	var ticker = time.NewTicker(duration)
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

			switch d.instance.Status.Gear {
			case v1alpha1.DummySpecialDeviceGearFast:
				if d.instance.Status.RotatingSpeed < 300 {
					d.instance.Status.RotatingSpeed++
				}
			case v1alpha1.DummySpecialDeviceGearMiddle:
				if d.instance.Status.RotatingSpeed < 200 {
					d.instance.Status.RotatingSpeed++
				}
			case v1alpha1.DummySpecialDeviceGearSlow:
				if d.instance.Status.RotatingSpeed < 100 {
					d.instance.Status.RotatingSpeed++
				}
			}
			d.sync()
		}()

		select {
		case <-stop:
			return
		default:
		}
	}
}

func (d *specialDevice) sync() {
	if d.toLimb != nil {
		d.toLimb(d.instance)
	}
	if d.mqttClient != nil {
		// NB(thxCode) we don't need to send extension status outside.
		var status = d.instance.Status.DeepCopy()
		status.Extension = nil
		if err := d.mqttClient.Publish(status); err != nil {
			d.log.Error(err, "Failed to publish MQTT broker")
		}
	}
}

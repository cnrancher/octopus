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
	"github.com/rancher/octopus/pkg/util/object"
)

func NewSpecialDevice(log logr.Logger, meta metav1.ObjectMeta, toLimb DummySpecialDeviceLimbSyncer) Device {
	log.Info("Created ")
	return &specialDevice{
		log: log,
		instance: &v1alpha1.DummySpecialDevice{
			ObjectMeta: meta,
		},
		toLimb: toLimb,
	}
}

type specialDevice struct {
	sync.Mutex

	log      logr.Logger
	instance *v1alpha1.DummySpecialDevice
	toLimb   DummySpecialDeviceLimbSyncer
	stop     chan struct{}

	mqttClient mqtt.Client
}

func (d *specialDevice) Configure(references api.ReferencesHandler, configuration interface{}) error {
	d.Lock()
	defer d.Unlock()

	var device, ok = configuration.(*v1alpha1.DummySpecialDevice)
	if !ok {
		d.log.Error(errors.New("invalidate configuration type"), "Failed to configure")
		return nil
	}
	var newSpec = device.Spec

	// configures MQTT client if needed
	var staleExtension, newExtension v1alpha1.DummyDeviceExtension
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

func (d *specialDevice) Shutdown() {
	d.Lock()
	defer d.Unlock()

	d.stopMock()
	if d.mqttClient != nil {
		d.mqttClient.Disconnect()
		d.mqttClient = nil
	}
	d.log.Info("Shutdown")
}

// refresh refreshes the status with new spec.
func (d *specialDevice) refresh(newSpec v1alpha1.DummySpecialDeviceSpec) error {
	var status = d.instance.Status
	if newSpec.On {
		var staleSpec = d.instance.Spec
		if staleSpec.Gear != newSpec.Gear {
			d.stopMock()

			status.Gear = newSpec.Gear
			switch status.Gear {
			case v1alpha1.DummySpecialDeviceGearFast:
				status.RotatingSpeed = 200
			case v1alpha1.DummySpecialDeviceGearMiddle:
				status.RotatingSpeed = 100
			case v1alpha1.DummySpecialDeviceGearSlow:
				status.RotatingSpeed = 0
			}
		}
		d.startMock(newSpec.Gear)
	} else {
		d.stopMock()
	}

	// records
	d.instance.Spec = newSpec
	d.instance.Status = status
	return d.sync()
}

// mock is blocked, it is used to simulate real device state changes
// and synchronize the changed values back to the limb.
func (d *specialDevice) mock(gear v1alpha1.DummySpecialDeviceGear, stop <-chan struct{}) {
	d.log.Info("Mocking")
	defer func() {
		d.log.Info("Finished mocking")
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

			var status = &d.instance.Status
			switch status.Gear {
			case v1alpha1.DummySpecialDeviceGearFast:
				if status.RotatingSpeed < 300 {
					status.RotatingSpeed++
				}
			case v1alpha1.DummySpecialDeviceGearMiddle:
				if status.RotatingSpeed < 200 {
					status.RotatingSpeed++
				}
			case v1alpha1.DummySpecialDeviceGearSlow:
				if status.RotatingSpeed < 100 {
					status.RotatingSpeed++
				}
			}
			if err := d.sync(); err != nil {
				d.log.Error(err, "failed to sync")
			}
		}()

		select {
		case <-stop:
			return
		default:
		}
	}
}

func (d *specialDevice) stopMock() {
	if d.stop != nil {
		close(d.stop)
		d.stop = nil
	}
}

func (d *specialDevice) startMock(gear v1alpha1.DummySpecialDeviceGear) {
	if d.stop == nil {
		d.stop = make(chan struct{})
		go d.mock(gear, d.stop)
	}
}

// sync combines all synchronization operations.
func (d *specialDevice) sync() error {
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

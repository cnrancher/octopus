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

func NewProtocolDevice(log logr.Logger, meta metav1.ObjectMeta, toLimb DummyProtocolDeviceLimbSyncer) Device {
	log.Info("Created ")
	return &protocolDevice{
		log: log,
		instance: &v1alpha1.DummyProtocolDevice{
			ObjectMeta: meta,
		},
		toLimb: toLimb,
	}
}

type protocolDevice struct {
	sync.Once
	sync.Mutex

	log      logr.Logger
	instance *v1alpha1.DummyProtocolDevice
	toLimb   DummyProtocolDeviceLimbSyncer
	stop     chan struct{}

	mqttClient mqtt.Client
}

func (d *protocolDevice) Configure(references api.ReferencesHandler, configuration interface{}) error {
	d.Lock()
	defer d.Unlock()

	var device, ok = configuration.(*v1alpha1.DummyProtocolDevice)
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

func (d *protocolDevice) Shutdown() {
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
func (d *protocolDevice) refresh(newSpec v1alpha1.DummyProtocolDeviceSpec) error {
	var status = d.instance.Status
	var staleSpec = d.instance.Spec
	if !reflect.DeepEqual(staleSpec.Properties, newSpec.Properties) {
		d.stopMock()

		status.Properties = make(map[string]v1alpha1.DummyProtocolDeviceStatusProperty, len(newSpec.Properties))
		fillStatusObject(newSpec.Properties, status.Properties)
	}

	// mocks in backend
	d.startMock()

	// records
	d.instance.Spec = newSpec
	d.instance.Status = status
	return d.sync()
}

// mock is blocked, it is used to simulate real device state changes
// and synchronize the changed values back to the limb.
func (d *protocolDevice) mock(stop <-chan struct{}) {
	d.log.Info("Mocking")
	defer func() {
		d.log.Info("Finished mocking")
	}()

	var ticker = time.NewTicker(2 * time.Second)
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

			fillStatusObject(d.instance.Spec.Properties, d.instance.Status.Properties)
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

func (d *protocolDevice) stopMock() {
	if d.stop != nil {
		close(d.stop)
		d.stop = nil
	}
}

func (d *protocolDevice) startMock() {
	if d.stop == nil {
		d.stop = make(chan struct{})
		go d.mock(d.stop)
	}
}

// sync combines all synchronization operations.
func (d *protocolDevice) sync() error {
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

func fillStatusArray(source v1alpha1.DummyProtocolDeviceProperty, length int) []v1alpha1.DummyProtocolDeviceStatusProperty {
	var target []v1alpha1.DummyProtocolDeviceStatusProperty
	var sourceProp = source
	for i := 0; i < length; i++ {
		switch sourceProp.Type {
		case v1alpha1.DummyProtocolDevicePropertyTypeBoolean:
			target = append(target, v1alpha1.DummyProtocolDeviceStatusProperty{
				Type:         v1alpha1.DummyProtocolDevicePropertyTypeBoolean,
				BooleanValue: randomBoolean(),
			})
		case v1alpha1.DummyProtocolDevicePropertyTypeFloat:
			target = append(target, v1alpha1.DummyProtocolDeviceStatusProperty{
				Type:       v1alpha1.DummyProtocolDevicePropertyTypeFloat,
				FloatValue: randomFloat(),
			})
		case v1alpha1.DummyProtocolDevicePropertyTypeInt:
			target = append(target, v1alpha1.DummyProtocolDeviceStatusProperty{
				Type:     v1alpha1.DummyProtocolDevicePropertyTypeInt,
				IntValue: randomInt(1000),
			})
		case v1alpha1.DummyProtocolDevicePropertyTypeString:
			target = append(target, v1alpha1.DummyProtocolDeviceStatusProperty{
				Type:        v1alpha1.DummyProtocolDevicePropertyTypeString,
				StringValue: randomString(10),
			})
		case v1alpha1.DummyProtocolDevicePropertyTypeArray:
			if sourceProp.ArrayProperties != nil {
				var items = fillStatusArray(sourceProp.ArrayProperties.DummyProtocolDeviceProperty, *randomInt(10))
				target = append(target, v1alpha1.DummyProtocolDeviceStatusProperty{
					Type:       v1alpha1.DummyProtocolDevicePropertyTypeArray,
					ArrayValue: toStatusObjectOrArrayPropsArray(items),
				})
			}
		case v1alpha1.DummyProtocolDevicePropertyTypeObject:
			if len(sourceProp.ObjectProperties) != 0 {
				var obj = make(map[string]v1alpha1.DummyProtocolDeviceStatusProperty, len(sourceProp.ObjectProperties))
				fillStatusObject(toSpecPropsObject(sourceProp.ObjectProperties), obj)
				target = append(target, v1alpha1.DummyProtocolDeviceStatusProperty{
					Type:        v1alpha1.DummyProtocolDevicePropertyTypeObject,
					ObjectValue: toStatusObjectOrArrayPropsObject(obj),
				})
			}
		}
	}

	return target
}

func fillStatusObject(source map[string]v1alpha1.DummyProtocolDeviceProperty, target map[string]v1alpha1.DummyProtocolDeviceStatusProperty) {
	for sourcePropName, sourceProp := range source {
		switch sourceProp.Type {
		case v1alpha1.DummyProtocolDevicePropertyTypeBoolean:
			target[sourcePropName] = v1alpha1.DummyProtocolDeviceStatusProperty{
				Type:         v1alpha1.DummyProtocolDevicePropertyTypeBoolean,
				BooleanValue: randomBoolean(),
			}
		case v1alpha1.DummyProtocolDevicePropertyTypeFloat:
			target[sourcePropName] = v1alpha1.DummyProtocolDeviceStatusProperty{
				Type:       v1alpha1.DummyProtocolDevicePropertyTypeFloat,
				FloatValue: randomFloat(),
			}
		case v1alpha1.DummyProtocolDevicePropertyTypeInt:
			target[sourcePropName] = v1alpha1.DummyProtocolDeviceStatusProperty{
				Type:     v1alpha1.DummyProtocolDevicePropertyTypeInt,
				IntValue: randomInt(1000),
			}
		case v1alpha1.DummyProtocolDevicePropertyTypeString:
			target[sourcePropName] = v1alpha1.DummyProtocolDeviceStatusProperty{
				Type:        v1alpha1.DummyProtocolDevicePropertyTypeString,
				StringValue: randomString(10),
			}
		case v1alpha1.DummyProtocolDevicePropertyTypeArray:
			if sourceProp.ArrayProperties != nil {
				var items = fillStatusArray(sourceProp.ArrayProperties.DummyProtocolDeviceProperty, *randomInt(10))
				target[sourcePropName] = v1alpha1.DummyProtocolDeviceStatusProperty{
					Type:       v1alpha1.DummyProtocolDevicePropertyTypeArray,
					ArrayValue: toStatusObjectOrArrayPropsArray(items),
				}
			}
		case v1alpha1.DummyProtocolDevicePropertyTypeObject:
			if len(sourceProp.ObjectProperties) != 0 {
				var obj = make(map[string]v1alpha1.DummyProtocolDeviceStatusProperty, len(sourceProp.ObjectProperties))
				fillStatusObject(toSpecPropsObject(sourceProp.ObjectProperties), obj)
				target[sourcePropName] = v1alpha1.DummyProtocolDeviceStatusProperty{
					Type:        v1alpha1.DummyProtocolDevicePropertyTypeObject,
					ObjectValue: toStatusObjectOrArrayPropsObject(obj),
				}
			}
		}
	}
}

func toStatusObjectOrArrayPropsArray(props []v1alpha1.DummyProtocolDeviceStatusProperty) []v1alpha1.DummyProtocolDeviceStatusObjectOrArrayProperty {
	var ret = make([]v1alpha1.DummyProtocolDeviceStatusObjectOrArrayProperty, 0, len(props))
	for _, prop := range props {
		ret = append(ret, v1alpha1.DummyProtocolDeviceStatusObjectOrArrayProperty{
			DummyProtocolDeviceStatusProperty: prop,
		})
	}
	return ret
}

func toStatusObjectOrArrayPropsObject(props map[string]v1alpha1.DummyProtocolDeviceStatusProperty) map[string]v1alpha1.DummyProtocolDeviceStatusObjectOrArrayProperty {
	var ret = make(map[string]v1alpha1.DummyProtocolDeviceStatusObjectOrArrayProperty, len(props))
	for propName, prop := range props {
		ret[propName] = v1alpha1.DummyProtocolDeviceStatusObjectOrArrayProperty{
			DummyProtocolDeviceStatusProperty: prop,
		}
	}
	return ret
}

func toSpecPropsObject(props map[string]v1alpha1.DummyProtocolDeviceObjectOrArrayProperty) map[string]v1alpha1.DummyProtocolDeviceProperty {
	var ret = make(map[string]v1alpha1.DummyProtocolDeviceProperty, len(props))
	for propName, prop := range props {
		ret[propName] = prop.DummyProtocolDeviceProperty
	}
	return ret
}

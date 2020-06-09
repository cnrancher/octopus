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

func NewProtocolDevice(log logr.Logger, instance *v1alpha1.DummyProtocolDevice, toLimb ProtocolDeviceSyncer) Device {
	return &protocolDevice{
		log: log,
		instance: &v1alpha1.DummyProtocolDevice{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: instance.Namespace,
				Name:      instance.Name,
				UID:       instance.UID,
			},
		},
		toLimb: toLimb,
	}
}

type protocolDevice struct {
	sync.Once
	sync.Mutex

	stop chan struct{}
	log  logr.Logger

	instance   *v1alpha1.DummyProtocolDevice
	toLimb     ProtocolDeviceSyncer
	mqttClient mqtt.Client
}

func (d *protocolDevice) Configure(references api.ReferencesHandler, configuration interface{}) error {
	var spec, ok = configuration.(v1alpha1.DummyProtocolDeviceSpec)
	if !ok {
		d.log.Error(errors.New("invalidate configuration type"), "Failed to configure")
		return nil
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
	shuffleStatus(d.instance)
	d.sync()

	d.Do(func() {
		d.stop = make(chan struct{})
		go d.mockPhysicalWatching(d.stop)
	})

	return nil
}

func (d *protocolDevice) Shutdown() {
	d.Lock()
	defer d.Unlock()

	if d.mqttClient != nil {
		d.mqttClient.Disconnect(5 * time.Second)
		d.mqttClient = nil
	}

	if d.stop != nil {
		close(d.stop)
		d.stop = nil
	}

	d.log.Info("Closed connection")
}

// mockPhysicalWatching is used to simulate real device state changes
// and synchronize the changed values back to the limb.
func (d *protocolDevice) mockPhysicalWatching(stop <-chan struct{}) {
	d.log.Info("Mocking started")
	defer func() {
		d.log.Info("Mocking finished")
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

			shuffleStatus(d.instance)
			d.sync()
		}()

		select {
		case <-stop:
			return
		default:
		}
	}
}

func (d *protocolDevice) sync() {
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

// Randomly generate some observed properties according to the desired properties.
func shuffleStatus(instance *v1alpha1.DummyProtocolDevice) {
	var statusProps = instance.Status.Props
	if len(statusProps) == 0 {
		statusProps = make(map[string]v1alpha1.DummyProtocolDeviceStatusProps, len(instance.Spec.Props))
	}

	fillStatusObject(instance.Spec.Props, statusProps)
	instance.Status.Props = statusProps
}

func fillStatusArray(source v1alpha1.DummyProtocolDeviceSpecProps, length int) []v1alpha1.DummyProtocolDeviceStatusProps {
	var target []v1alpha1.DummyProtocolDeviceStatusProps
	var sourceProp = source
	for i := 0; i < length; i++ {
		switch sourceProp.Type {
		case v1alpha1.DummyProtocolDevicePropertyTypeBoolean:
			target = append(target, v1alpha1.DummyProtocolDeviceStatusProps{
				Type:         v1alpha1.DummyProtocolDevicePropertyTypeBoolean,
				BooleanValue: randomBoolean(),
			})
		case v1alpha1.DummyProtocolDevicePropertyTypeFloat:
			target = append(target, v1alpha1.DummyProtocolDeviceStatusProps{
				Type:       v1alpha1.DummyProtocolDevicePropertyTypeFloat,
				FloatValue: randomFloat(),
			})
		case v1alpha1.DummyProtocolDevicePropertyTypeInt:
			target = append(target, v1alpha1.DummyProtocolDeviceStatusProps{
				Type:     v1alpha1.DummyProtocolDevicePropertyTypeInt,
				IntValue: randomInt(1000),
			})
		case v1alpha1.DummyProtocolDevicePropertyTypeString:
			target = append(target, v1alpha1.DummyProtocolDeviceStatusProps{
				Type:        v1alpha1.DummyProtocolDevicePropertyTypeString,
				StringValue: randomString(10),
			})
		case v1alpha1.DummyProtocolDevicePropertyTypeArray:
			if sourceProp.ArrayProps != nil {
				var items = fillStatusArray(sourceProp.ArrayProps.DummyProtocolDeviceSpecProps, *randomInt(10))
				target = append(target, v1alpha1.DummyProtocolDeviceStatusProps{
					Type:       v1alpha1.DummyProtocolDevicePropertyTypeArray,
					ArrayValue: toStatusObjectOrArrayPropsArray(items),
				})
			}
		case v1alpha1.DummyProtocolDevicePropertyTypeObject:
			if len(sourceProp.ObjectProps) != 0 {
				var object = make(map[string]v1alpha1.DummyProtocolDeviceStatusProps, len(sourceProp.ObjectProps))
				fillStatusObject(toSpecPropsObject(sourceProp.ObjectProps), object)
				target = append(target, v1alpha1.DummyProtocolDeviceStatusProps{
					Type:        v1alpha1.DummyProtocolDevicePropertyTypeObject,
					ObjectValue: toStatusObjectOrArrayPropsObject(object),
				})
			}
		}
	}

	return target
}

func fillStatusObject(source map[string]v1alpha1.DummyProtocolDeviceSpecProps, target map[string]v1alpha1.DummyProtocolDeviceStatusProps) {
	for sourcePropName, sourceProp := range source {
		if _, exist := target[sourcePropName]; exist && sourceProp.ReadOnly {
			continue
		}

		switch sourceProp.Type {
		case v1alpha1.DummyProtocolDevicePropertyTypeBoolean:
			target[sourcePropName] = v1alpha1.DummyProtocolDeviceStatusProps{
				Type:         v1alpha1.DummyProtocolDevicePropertyTypeBoolean,
				BooleanValue: randomBoolean(),
			}
		case v1alpha1.DummyProtocolDevicePropertyTypeFloat:
			target[sourcePropName] = v1alpha1.DummyProtocolDeviceStatusProps{
				Type:       v1alpha1.DummyProtocolDevicePropertyTypeFloat,
				FloatValue: randomFloat(),
			}
		case v1alpha1.DummyProtocolDevicePropertyTypeInt:
			target[sourcePropName] = v1alpha1.DummyProtocolDeviceStatusProps{
				Type:     v1alpha1.DummyProtocolDevicePropertyTypeInt,
				IntValue: randomInt(1000),
			}
		case v1alpha1.DummyProtocolDevicePropertyTypeString:
			target[sourcePropName] = v1alpha1.DummyProtocolDeviceStatusProps{
				Type:        v1alpha1.DummyProtocolDevicePropertyTypeString,
				StringValue: randomString(10),
			}
		case v1alpha1.DummyProtocolDevicePropertyTypeArray:
			if sourceProp.ArrayProps != nil {
				var items = fillStatusArray(sourceProp.ArrayProps.DummyProtocolDeviceSpecProps, *randomInt(10))
				target[sourcePropName] = v1alpha1.DummyProtocolDeviceStatusProps{
					Type:       v1alpha1.DummyProtocolDevicePropertyTypeArray,
					ArrayValue: toStatusObjectOrArrayPropsArray(items),
				}
			}
		case v1alpha1.DummyProtocolDevicePropertyTypeObject:
			if len(sourceProp.ObjectProps) != 0 {
				var object = make(map[string]v1alpha1.DummyProtocolDeviceStatusProps, len(sourceProp.ObjectProps))
				fillStatusObject(toSpecPropsObject(sourceProp.ObjectProps), object)
				target[sourcePropName] = v1alpha1.DummyProtocolDeviceStatusProps{
					Type:        v1alpha1.DummyProtocolDevicePropertyTypeObject,
					ObjectValue: toStatusObjectOrArrayPropsObject(object),
				}
			}
		}
	}
}

func toStatusObjectOrArrayPropsArray(props []v1alpha1.DummyProtocolDeviceStatusProps) []v1alpha1.DummyProtocolDeviceStatusObjectOrArrayProps {
	var ret = make([]v1alpha1.DummyProtocolDeviceStatusObjectOrArrayProps, 0, len(props))
	for _, prop := range props {
		ret = append(ret, v1alpha1.DummyProtocolDeviceStatusObjectOrArrayProps{
			DummyProtocolDeviceStatusProps: prop,
		})
	}
	return ret
}

func toStatusObjectOrArrayPropsObject(props map[string]v1alpha1.DummyProtocolDeviceStatusProps) map[string]v1alpha1.DummyProtocolDeviceStatusObjectOrArrayProps {
	var ret = make(map[string]v1alpha1.DummyProtocolDeviceStatusObjectOrArrayProps, len(props))
	for propName, prop := range props {
		ret[propName] = v1alpha1.DummyProtocolDeviceStatusObjectOrArrayProps{
			DummyProtocolDeviceStatusProps: prop,
		}
	}
	return ret
}

func toSpecPropsObject(props map[string]v1alpha1.DummyProtocolDeviceSpecObjectOrArrayProps) map[string]v1alpha1.DummyProtocolDeviceSpecProps {
	var ret = make(map[string]v1alpha1.DummyProtocolDeviceSpecProps, len(props))
	for propName, prop := range props {
		ret[propName] = prop.DummyProtocolDeviceSpecProps
	}
	return ret
}

package physical

import (
	"reflect"
	"sync"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/mqtt"
	"github.com/rancher/octopus/pkg/util/converter"
	"github.com/rancher/octopus/pkg/util/object"
)

// Device is an interface for device operations set.
type Device interface {
	// Shutdown uses to close the connection between adaptor and real(physical) device.
	Shutdown()
	// Configure uses to set up the device.
	Configure(references api.ReferencesHandler, configuration interface{}) error
}

// NewDevice creates a Device.
func NewDevice(log logr.Logger, meta metav1.ObjectMeta, toLimb MQTTDeviceLimbSyncer) Device {
	log.Info("Created")
	return &mqttDevice{
		log: log,
		instance: &v1alpha1.MQTTDevice{
			ObjectMeta: meta,
		},
		toLimb: toLimb,
	}
}

type mqttDevice struct {
	sync.Mutex

	log logr.Logger

	instance   *v1alpha1.MQTTDevice
	toLimb     MQTTDeviceLimbSyncer
	mqttClient mqtt.Client
}

func (d *mqttDevice) Shutdown() {
	d.log.Info("Shutdown")
	if d.mqttClient != nil {
		d.mqttClient.Disconnect()
		d.mqttClient = nil
		d.log.V(1).Info("Disconnected connection")
	}
}

func (d *mqttDevice) Configure(references api.ReferencesHandler, configuration interface{}) error {
	var device, ok = configuration.(*v1alpha1.MQTTDevice)
	if !ok {
		d.log.Error(errors.New("invalidate configuration type"), "Failed to configure")
		return nil
	}
	var newSpec = device.Spec

	d.Lock()
	defer d.Unlock()

	if !reflect.DeepEqual(d.instance.Spec.Protocol, newSpec.Protocol) {
		if d.mqttClient != nil {
			d.mqttClient.Disconnect()
			d.mqttClient = nil
			d.log.V(1).Info("Disconnected stale connection")
		}

		var cli, err = mqtt.NewClient(newSpec.Protocol.MQTTOptions, object.GetControlledOwnerObjectReference(device), references)
		if err != nil {
			return errors.Wrap(err, "failed to create MQTT client")
		}

		err = cli.Connect()
		if err != nil {
			return errors.Wrap(err, "failed to connect MQTT broker")
		}
		d.mqttClient = cli
		d.log.V(1).Info("Connected to MQTT broker")
	}

	return d.refresh(newSpec)
}

func (d *mqttDevice) refresh(newSpec v1alpha1.MQTTDeviceSpec) error {
	// indexes stale status properties
	var staleStatusPropsIndex = make(map[string]v1alpha1.MQTTDeviceStatusProperty, len(d.instance.Status.Properties))
	for _, prop := range d.instance.Status.Properties {
		staleStatusPropsIndex[prop.Name] = prop
	}

	// constructs status properties
	var newStatusProps = make([]v1alpha1.MQTTDeviceStatusProperty, 0, len(newSpec.Properties))
	for _, newSpecProp := range newSpec.Properties {
		switch newSpec.Protocol.Pattern {
		case v1alpha1.MQTTDevicePatternAttributedMessage:
			if newSpecProp.ReadOnly != nil && !*newSpecProp.ReadOnly {
				if err := verifyWritableJSONPath(getPath(newSpecProp.Name, newSpecProp.Path)); err != nil {
					return errors.Wrapf(err, "illegal path %s", getPath(newSpecProp.Name, newSpecProp.Path))
				}
			}
		}

		var newStatusProp = v1alpha1.MQTTDeviceStatusProperty{
			MQTTDeviceProperty: newSpecProp,
		}
		var staleStatusProp, exist = staleStatusPropsIndex[newSpecProp.Name]
		if !exist {
			newStatusProp.Value = nil
		} else {
			newStatusProp.Value = staleStatusProp.Value
			newStatusProp.UpdatedAt = staleStatusProp.UpdatedAt
		}
		newStatusProps = append(newStatusProps, newStatusProp)
	}

	// indexes stale spec properties
	var staleSpecPropsIndex = make(map[string]v1alpha1.MQTTDeviceProperty, len(d.instance.Spec.Properties))
	for _, prop := range d.instance.Spec.Properties {
		staleSpecPropsIndex[prop.Name] = prop
	}

	// refreshes
	switch newSpec.Protocol.Pattern {
	case v1alpha1.MQTTDevicePatternAttributedMessage:
		if err := d.refreshAsAttributedMessage(staleSpecPropsIndex, newSpec.Properties); err != nil {
			return err
		}
	case v1alpha1.MQTTDevicePatternAttributeTopic:
		if err := d.refreshAsAttributedTopic(staleSpecPropsIndex, newSpec.Properties); err != nil {
			return err
		}
	default:
		return errors.Errorf("failed to recognize protocol pattern %s", newSpec.Protocol.Pattern)
	}

	// records
	d.instance.Spec = newSpec
	d.instance.Status = v1alpha1.MQTTDeviceStatus{Properties: newStatusProps}
	d.toLimb(d.instance)
	return nil
}

func (d *mqttDevice) refreshAsAttributedMessage(staleSpecPropsIndex map[string]v1alpha1.MQTTDeviceProperty, newSpecProps []v1alpha1.MQTTDeviceProperty) error {
	// subscribes
	var subscribeTopics = []mqtt.SubscribeTopic{{}}
	var subscribeHandler = func(msg mqtt.SubscribeMessage) {
		// receives and updates status properties
		d.Lock()
		defer d.Unlock()

		var payload = msg.Payload
		for idx, prop := range d.instance.Status.Properties {
			var propValue = &v1alpha1.MQTTDevicePropertyValue{}
			var result = gjson.GetBytes(payload, getPath(prop.Name, prop.Path))
			if result.Index > 0 {
				propValue.Raw = payload[result.Index : result.Index+len(result.Raw)]
			} else {
				propValue.Raw = []byte(result.Raw)
			}
			prop.Value = propValue
			prop.UpdatedAt = now()
			d.instance.Status.Properties[idx] = prop
		}
		d.toLimb(d.instance)
	}
	if err := d.mqttClient.Subscribe(subscribeTopics, subscribeHandler); err != nil {
		return errors.Wrap(err, "failed to subscribe")
	}

	// publishes
	var stalePayload []byte
	var payload []byte
	for _, newSpecProp := range newSpecProps {
		if newSpecProp.ReadOnly != nil && !*newSpecProp.ReadOnly {
			// constructs stale payload
			if staleSpecProp, exist := staleSpecPropsIndex[newSpecProp.Name]; exist {
				if staleSpecProp.Value != nil {
					var stalePropPath = getPath(staleSpecProp.Name, staleSpecProp.Path)
					stalePayload, _ = sjson.SetBytes(payload, stalePropPath, staleSpecProp.Value)
				}
			}

			// constructs new payload
			if newSpecProp.Value != nil {
				var newPropPath = getPath(newSpecProp.Name, newSpecProp.Path)
				var err error
				payload, err = sjson.SetBytes(payload, newPropPath, newSpecProp.Value)
				if err != nil {
					return errors.Wrapf(err, "failed to set property value on path: %s", newPropPath)
				}
			}
		}
	}
	if !reflect.DeepEqual(stalePayload, payload) {
		if err := d.mqttClient.Publish(mqtt.PublishMessage{Payload: payload}); err != nil {
			return errors.Wrap(err, "failed to publish")
		}
	}

	return nil
}

func (d *mqttDevice) refreshAsAttributedTopic(staleSpecPropsIndex map[string]v1alpha1.MQTTDeviceProperty, newSpecProps []v1alpha1.MQTTDeviceProperty) error {
	// subscribes spec properties
	var subscribeTopics = make([]mqtt.SubscribeTopic, 0, len(newSpecProps))
	for idx, newSpecProp := range newSpecProps {
		// appends subscribe topic
		subscribeTopics = append(subscribeTopics, mqtt.SubscribeTopic{
			Index:      idx,
			Render:     getSubscribeRender(&newSpecProp),
			QoSPointer: (*byte)(newSpecProp.QoS),
		})
	}
	var subscribeHandler = func(msg mqtt.SubscribeMessage) {
		// receives and updates status properties
		d.Lock()
		defer d.Unlock()

		if msg.Index > len(d.instance.Status.Properties) {
			return
		}

		var propValue v1alpha1.MQTTDevicePropertyValue
		if err := converter.UnmarshalJSON(msg.Payload, &propValue); err != nil {
			d.log.Error(err, "Failed to unmarshal subscribed payload", "topic", msg.Topic)
			return
		}
		var prop = &d.instance.Status.Properties[msg.Index]
		prop.Value = &propValue
		prop.UpdatedAt = now()
		// TODO should we debounce here?
		d.toLimb(d.instance)
	}
	if err := d.mqttClient.Subscribe(subscribeTopics, subscribeHandler); err != nil {
		return errors.Wrap(err, "failed to subscribe")
	}

	// publishes writable spec properties
	for _, newSpecProp := range newSpecProps {
		if newSpecProp.ReadOnly != nil && !*newSpecProp.ReadOnly {
			var staleSpecProp = staleSpecPropsIndex[newSpecProp.Name]
			// publishes again if changed
			if !reflect.DeepEqual(staleSpecProp, newSpecProp) {
				var err = d.mqttClient.Publish(mqtt.PublishMessage{
					Render:          getPublishRender(&newSpecProp),
					QoSPointer:      (*byte)(newSpecProp.QoS),
					RetainedPointer: newSpecProp.Retained,
					Payload:         newSpecProp.Value,
				})
				if err != nil {
					return errors.Wrapf(err, "failed to publish property %s", newSpecProp.Name)
				}
			}
		}
	}

	return nil
}

func getPath(name, path string) string {
	if path != "" {
		return path
	}
	return name
}

func getPublishRender(prop *v1alpha1.MQTTDeviceProperty) map[string]string {
	var render = make(map[string]string, 2)

	// gets path rendering value
	render["path"] = getPath(prop.Name, prop.Path)

	// gets operator rendering value
	if prop.Operator != nil {
		render["operator"] = prop.Operator.Write
	}

	return render
}

func getSubscribeRender(prop *v1alpha1.MQTTDeviceProperty) map[string]string {
	var render = make(map[string]string, 2)

	// gets path rendering value
	render["path"] = getPath(prop.Name, prop.Path)

	// gets operator rendering value
	if prop.Operator != nil {
		render["operator"] = prop.Operator.Read
	}

	return render
}

func now() *metav1.Time {
	var ret = metav1.Now()
	return &ret
}

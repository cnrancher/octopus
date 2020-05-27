package physical

import (
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-logr/logr"
	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
	"github.com/rancher/octopus/pkg/util/object"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	disconnectQuiesce = 1000
	waitTimeout       = time.Second * 10
)

type Device interface {
	Configure(spec *v1alpha1.MqttDeviceSpec)
	On()
	Shutdown()
}

func NewDevice(log logr.Logger, obj *v1alpha1.MqttDevice, handler DataHandler, client MQTT.Client) Device {
	d := device{
		client:  client,
		handler: handler,
		stop:    make(chan struct{}),
		log:     log,
	}
	obj.DeepCopyInto(&d.obj)

	return &d
}

type device struct {
	sync.Mutex
	log        logr.Logger
	client     MQTT.Client
	handler    DataHandler
	obj        v1alpha1.MqttDevice
	stop       chan struct{}
	payloadMap sync.Map
	o          sync.Once
}

func (dev *device) On() {
	if err := dev.subscribe(); err != nil {
		dev.log.Error(err, "device subscribe error")
		close(dev.stop)
		return
	}

	select {
	case <-dev.stop:
		dev.log.Info("device is stop")
		return
	}
}

func (dev *device) Configure(spec *v1alpha1.MqttDeviceSpec) {
	dev.Lock()
	defer dev.Unlock()
	if err := dev.updateSubscription(spec); err != nil {
		dev.log.Error(err, "device Configure updateSubscription error")
		return
	}
	if err := dev.publishProperties(spec.Properties); err != nil {
		dev.log.Error(err, "device Configure publish error")
		return
	}
	dev.removeRedundantStatus(spec)
	dev.updateDeviceSpec(spec)
}

func (dev *device) Shutdown() {
	dev.o.Do(func() {
		close(dev.stop)
		dev.unsubscribeAll()
		dev.client.Disconnect(disconnectQuiesce)
	})
}

func (dev *device) publishProperties(properties []v1alpha1.Property) error {
	for _, property := range properties {
		if property.SubInfo.PayloadType != v1alpha1.PayloadTypeJSON {
			continue
		}
		var statusProperty v1alpha1.StatusProperty
		for _, sp := range dev.obj.Status.Properties {
			if sp.Name == property.Name {
				statusProperty = sp
				break
			}
		}
		if ComparativeValueProps(property.Value, statusProperty.Value) {
			continue
		}

		ivalue, ok := dev.payloadMap.Load(property.SubInfo.Topic)
		if !ok {
			continue
		}
		payload := ivalue.([]byte)
		newValuePayload, err := ConvertValueToJSONPayload(payload, &property)
		if err != nil {
			return err
		}

		var pubTopic string
		var qos byte
		if property.PubInfo.Topic == "" {
			pubTopic = property.SubInfo.Topic
			qos = byte(property.SubInfo.Qos)
		} else {
			pubTopic = property.PubInfo.Topic
			qos = byte(property.PubInfo.Qos)
		}

		dev.log.Info("device publish cmd", "payload", string(newValuePayload), "propertyName", property.Name, "pubTopic", pubTopic)

		token := dev.client.Publish(pubTopic, qos, true, newValuePayload)
		if token.WaitTimeout(waitTimeout) && token.Error() != nil {
			return err
		}

		dev.payloadMap.Store(property.SubInfo.Topic, newValuePayload)
	}
	return nil
}

func (dev *device) subscribe() error {
	filters := make(map[string]byte, len(dev.obj.Spec.Properties))
	for _, property := range dev.obj.Spec.Properties {
		filters[property.SubInfo.Topic] = byte(property.SubInfo.Qos)
	}

	token := dev.client.SubscribeMultiple(filters, dev.callback)
	if token.WaitTimeout(time.Second*10) && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (dev *device) callback(client MQTT.Client, msg MQTT.Message) {
	dev.Lock()
	defer dev.Unlock()
	dev.log.Info("device subscribe callback", "msg", string(msg.Payload()), "topic", msg.Topic())
	dev.log.Info("device current object", "object", dev.obj)
	dev.payloadMap.Store(msg.Topic(), msg.Payload())
	for _, property := range dev.obj.Spec.Properties {
		if property.SubInfo.Topic == msg.Topic() {
			statusProperty, err := ConvertToStatusProperty(msg.Payload(), &property)
			if err != nil {
				dev.log.Error(err, "device subscribe callback ConvertToStatusProperty error", "property", property)
				continue
			}

			dev.log.Info("device subscribe callback ConvertToStatusProperty", "statusProperty", statusProperty)

			var found bool
			for j, curStatusProperty := range dev.obj.Status.Properties {
				if curStatusProperty.Name == property.Name {
					statusProperty.DeepCopyInto(&dev.obj.Status.Properties[j])
					found = true
				}
			}
			if !found {
				dev.obj.Status.Properties = append(dev.obj.Status.Properties, statusProperty)
			}
		}
	}

	dev.handler(object.GetNamespacedName(&dev.obj), dev.obj.Status)
}

func (dev *device) updateSubscription(spec *v1alpha1.MqttDeviceSpec) error {
	if err := dev.unsubscribeOld(spec); err != nil {
		return err
	}
	if err := dev.reSubscribeAll(spec); err != nil {
		return err
	}

	return nil
}

func (dev *device) unsubscribeOld(spec *v1alpha1.MqttDeviceSpec) error {
	oldTopicSet := sets.NewString()
	for _, property := range dev.obj.Spec.Properties {
		oldTopicSet.Insert(property.SubInfo.Topic)
	}

	newTopicSet := sets.NewString()
	for _, property := range spec.Properties {
		newTopicSet.Insert(property.SubInfo.Topic)
	}

	allTopicSet := oldTopicSet.Union(newTopicSet)
	mustDelTopicSet := allTopicSet.Difference(newTopicSet)

	var unSubTopic []string
	for _, topic := range mustDelTopicSet.List() {
		unSubTopic = append(unSubTopic, topic)
	}

	if len(unSubTopic) > 1 {
		if token := dev.client.Unsubscribe(unSubTopic...); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}

	return nil
}

func (dev *device) reSubscribeAll(spec *v1alpha1.MqttDeviceSpec) error {
	dev.client.Disconnect(disconnectQuiesce)

	var err error
	if dev.client, err = NewMqttClient(dev.obj.Name, spec.Config); err != nil {
		return err
	}

	filters := make(map[string]byte, len(spec.Properties))
	for _, property := range spec.Properties {
		filters[property.SubInfo.Topic] = byte(property.SubInfo.Qos)
	}

	token := dev.client.SubscribeMultiple(filters, dev.callback)
	if token.WaitTimeout(waitTimeout) && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (dev *device) removeRedundantStatus(spec *v1alpha1.MqttDeviceSpec) {
	for i := 0; i < len(dev.obj.Status.Properties); i++ {
		statusProperty := dev.obj.Status.Properties[i]
		var found bool
		for _, property := range spec.Properties {
			if property.Name == statusProperty.Name {
				found = true
				break
			}
		}
		if !found {
			dev.obj.Status.Properties = append(dev.obj.Status.Properties[:i], dev.obj.Status.Properties[i+1:]...)
		}
	}
}

func (dev *device) unsubscribeAll() {
	subSet := sets.NewString()
	for _, property := range dev.obj.Spec.Properties {
		if subSet.HasAny(property.SubInfo.Topic) {
			continue
		}
		subSet.Insert(property.SubInfo.Topic)
	}
	token := dev.client.Unsubscribe(subSet.List()...)
	if token.WaitTimeout(waitTimeout) && token.Error() != nil {
		dev.log.Error(token.Error(), "device unsubscribeAll error")
	}
	return
}

func (dev *device) updateDeviceSpec(spec *v1alpha1.MqttDeviceSpec) {
	spec.DeepCopyInto(&dev.obj.Spec)
}

func NewMqttClient(clientID string, config v1alpha1.MqttConfig) (MQTT.Client, error) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.Broker)
	opts.SetClientID(clientID)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)
	opts.SetOrderMatters(true)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(false)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.WaitTimeout(waitTimeout) && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

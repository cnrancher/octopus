package physical

import (
	"io"
	"reflect"
	"sync"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
	"github.com/rancher/octopus/adaptors/mqtt/pkg/metadata"
	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/adaptor/socket/handler"
	"github.com/rancher/octopus/pkg/mqtt"
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
	log.Info("Created ")
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

	log        logr.Logger
	instance   *v1alpha1.MQTTDevice
	toLimb     MQTTDeviceLimbSyncer
	mqttClient mqtt.Client
}

func (d *mqttDevice) Shutdown() {
	d.Lock()
	defer d.Unlock()

	if d.mqttClient != nil {
		d.mqttClient.Disconnect()
		d.mqttClient = nil
		d.log.V(1).Info("Disconnected connection")
	}
	d.log.Info("Shutdown")
}

func (d *mqttDevice) Configure(references api.ReferencesHandler, configuration interface{}) error {
	defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

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

		var clientBuilder = mqtt.NewClientBuilder(newSpec.Protocol.MQTTOptions, object.GetControlledOwnerObjectReference(device))
		clientBuilder.Render(references)
		clientBuilder.ConfigureOptions(func(options *MQTT.ClientOptions) error {
			var autoReconnect = options.AutoReconnect
			options.SetConnectionLostHandler(func(_ MQTT.Client, cerr error) {
				if autoReconnect {
					d.log.Error(cerr, "MQTT broker connection is closed, please turn off the AutoReconnect if want to know this at the first time")
					return
				}

				// NB(thxCode) feedbacks the EOF of MQTT broker connection if turn off the auto reconnection.
				var feedbackErr error
				if cerr != io.EOF {
					feedbackErr = errors.Wrapf(cerr, "error for MQTT broker connection")
				} else {
					feedbackErr = errors.New("MQTT broker connection is closed")
				}
				if d.toLimb != nil {
					if err := d.toLimb(nil, feedbackErr); err != nil {
						d.log.Error(err, "failed to feedback the lost error of MQTT broker connection")
					}
				}
			})
			return nil
		})
		var cli, err = clientBuilder.Build()
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

// refresh refreshes the status with new spec.
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
	return d.sync()
}

// refreshAsAttributedMessage treats all properties as a whole JSON payload.
// When subscribing, the data in JSON will be obtained according to the `path` of each property.
// When publishing, all writable properties will be assembled into a JSON for transmission.
// It is worth noting that in order to reduce publishing,
// only when the value of the writable property changes will be pushed.
func (d *mqttDevice) refreshAsAttributedMessage(staleSpecPropsIndex map[string]v1alpha1.MQTTDeviceProperty, newSpecProps []v1alpha1.MQTTDeviceProperty) error {
	// subscribes
	var subscribeTopics = []mqtt.SubscribeTopic{{}}
	var subscribeHandler = func(msg mqtt.SubscribeMessage) {
		go func() {
			defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

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
			d.log.V(4).Info("Received payload", "type", "AttributedMessage")
			if err := d.sync(); err != nil {
				d.log.Error(err, "failed to sync")
			}
		}()
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
	if len(payload) != 0 && !reflect.DeepEqual(stalePayload, payload) {
		if err := d.mqttClient.Publish(mqtt.PublishMessage{Payload: payload}); err != nil {
			return errors.Wrap(err, "failed to publish")
		}
		d.log.V(4).Info("Sent payload", "type", "AttributedMessage")
	}

	return nil
}

// refreshAsAttributedTopic treats each property as a JSON payload.
// When subscribing, it will use the property `path` and `operator.read` to render the topic,
// and then subscribe to the rendered topic.
// When publishing, it will use the property `path` and `operator.write` to render the topic,
// and the publish to the rendered topic.
// It is worth noting that in order to reduce publishing,
// only when the value of the writable property changes will be pushed.
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
		go func() {
			defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

			// receives and updates status properties
			d.Lock()
			defer d.Unlock()

			if msg.Index > len(d.instance.Status.Properties) {
				return
			}

			var prop = &d.instance.Status.Properties[msg.Index]
			prop.Value = &v1alpha1.MQTTDevicePropertyValue{Raw: msg.Payload}
			prop.UpdatedAt = now()
			d.log.V(4).Info("Received payload", "type", "AttributedTopic", "property", prop.Name)
			// TODO should we debounce here?
			if err := d.sync(); err != nil {
				d.log.Error(err, "failed to sync")
			}
		}()
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
				d.log.V(4).Info("Sent payload", "type", "AttributedTopic", "property", newSpecProp.Name)
			}
		}
	}

	return nil
}

// sync combines all synchronization operations.
func (d *mqttDevice) sync() error {
	if d.toLimb != nil {
		if err := d.toLimb(d.instance, nil); err != nil {
			return err
		}
	}
	d.log.V(1).Info("Synced")
	return nil
}

// getPath returns the name as path if the path parameter is blank.
func getPath(name, path string) string {
	if path != "" {
		return path
	}
	return name
}

// getPublishRender returns the render map for published topic rendering.
// It is worth noting that the `operator.write: "null"` will be treated as blank string.
func getPublishRender(prop *v1alpha1.MQTTDeviceProperty) map[string]string {
	var render = make(map[string]string, 2)

	// gets path rendering value
	render["path"] = getPath(prop.Name, prop.Path)

	// gets operator rendering value
	if prop.Operator != nil {
		var write = prop.Operator.Write
		if write == "null" {
			write = ""
		}
		render["operator"] = write
	}

	return render
}

// getSubscribeRender returns the render map for subscribed topic rendering.
// It is worth noting that the `operator.read: "null"` will be treated as blank string.
func getSubscribeRender(prop *v1alpha1.MQTTDeviceProperty) map[string]string {
	var render = make(map[string]string, 2)

	// gets path rendering value
	render["path"] = getPath(prop.Name, prop.Path)

	// gets operator rendering value
	if prop.Operator != nil {
		var read = prop.Operator.Read
		if read == "null" {
			read = ""
		}
		render["operator"] = read
	}

	return render
}

func now() *metav1.Time {
	var ret = metav1.Now()
	return &ret
}

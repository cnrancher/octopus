package physical

import (
	"io"
	"reflect"
	"sync"
	"time"

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
	stop       chan struct{}
	mqttClient mqtt.Client
}

func (d *mqttDevice) Configure(references api.ReferencesHandler, configuration interface{}) error {
	defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

	var device, ok = configuration.(*v1alpha1.MQTTDevice)
	if !ok {
		d.log.Error(errors.New("invalidate configuration type"), "Failed to configure")
		return nil
	}

	d.Lock()
	defer d.Unlock()

	var newSpec = device.Spec
	var staleSpec = d.instance.Spec

	if !reflect.DeepEqual(staleSpec.Protocol, newSpec.Protocol) {
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

		// NB(thxCode) since the client has been changed,
		// we need to reset.
		d.instance.Spec = v1alpha1.MQTTDeviceSpec{}
	}

	// refreshes
	switch newSpec.Protocol.Pattern {
	case v1alpha1.MQTTDevicePatternAttributedMessage:
		return d.refreshAsAttributedMessage(newSpec)
	case v1alpha1.MQTTDevicePatternAttributeTopic:
		return d.refreshAsAttributedTopic(newSpec)
	}
	return errors.Errorf("failed to recognize protocol pattern %s", newSpec.Protocol.Pattern)
}

func (d *mqttDevice) Shutdown() {
	d.Lock()
	defer d.Unlock()

	d.stopFetch()
	if d.mqttClient != nil {
		d.mqttClient.Disconnect()
		d.mqttClient = nil
		d.log.V(1).Info("Disconnected connection")
	}
	d.log.Info("Shutdown")
}

// refreshAsAttributedMessage treats all properties as a whole JSON payload.
// When subscribing, the data in JSON will be obtained according to the `path` of each property.
// When publishing, all writable properties will be assembled into a JSON for transmission.
// It is worth noting that in order to reduce publishing,
// only when the value of the writable property changes will be pushed.
func (d *mqttDevice) refreshAsAttributedMessage(newSpec v1alpha1.MQTTDeviceSpec) error {
	var newStatus v1alpha1.MQTTDeviceStatus

	var staleStatus = d.instance.Status
	var staleSpec = d.instance.Spec
	if !reflect.DeepEqual(staleSpec.Properties, newSpec.Properties) {
		d.stopFetch()

		var staleSpecPropsMap = mapSpecProperties(staleSpec.Properties)
		var staleStatusPropsMap = mapStatusProperties(staleStatus.Properties)
		var writablePayload []byte

		// syncs properties
		var statusProps = make([]v1alpha1.MQTTDeviceStatusProperty, 0, len(newSpec.Properties))
		for i := 0; i < len(newSpec.Properties); i++ {
			var specPropPtr = &newSpec.Properties[i]
			var statusProp v1alpha1.MQTTDeviceStatusProperty
			if staleStatusPropPtr, existed := staleStatusPropsMap[specPropPtr.Name]; existed {
				statusProp = *staleStatusPropPtr
			} else {
				statusProp = v1alpha1.MQTTDeviceStatusProperty{
					Name:        specPropPtr.Name,
					Type:        specPropPtr.Type,
					AccessModes: specPropPtr.AccessModes,
					UpdatedAt:   now(),
				}
			}

			var err error
			for _, accessMode := range specPropPtr.MergeAccessModes() {
				switch accessMode {
				case v1alpha1.MQTTDevicePropertyAccessModeWriteOnce:
					if !reflect.DeepEqual(specPropPtr, staleSpecPropsMap[specPropPtr.Name]) {
						writablePayload, err = d.constructAttributedMessagePayload(writablePayload, specPropPtr)
						if err != nil {
							return errors.Wrapf(err, "failed to write property %s", specPropPtr.Name)
						}
					}
				case v1alpha1.MQTTDevicePropertyAccessModeWriteMany:
					writablePayload, err = d.constructAttributedMessagePayload(writablePayload, specPropPtr)
					if err != nil {
						return errors.Wrapf(err, "failed to write property %s", specPropPtr.Name)
					}
				}
				// NB(thxCode) MergeAccessModes has already processed the accessModes,
				// and in this loop, we only need to process the "Write*" property.
				break
			}

			statusProps = append(statusProps, statusProp)
		}

		// publishes to MQTT broker
		if len(writablePayload) != 0 {
			if err := d.mqttClient.Publish(mqtt.PublishMessage{Payload: writablePayload}); err != nil {
				return err
			}
			d.log.V(4).Info("Write properties", "payload", string(writablePayload))
		}

		// subscribes to MQTT broker
		if err := d.mqttClient.Subscribe(mqtt.SubscribeTopic{}); err != nil {
			return err
		}

		newStatus = v1alpha1.MQTTDeviceStatus{Properties: statusProps}
	} else {
		newStatus = staleStatus
	}

	// fetches in backend
	if err := d.startFetch(newSpec.Protocol.GetSyncInterval()); err != nil {
		return errors.Wrap(err, "failed to start fetch")
	}

	// records
	d.instance.Spec = newSpec
	d.instance.Status = newStatus
	return d.sync()
}

// refreshAsAttributedTopic treats each property as a JSON payload.
// When subscribing, it will use the property `path` and `operator.read` to render the topic,
// and then subscribe to the rendered topic.
// When publishing, it will use the property `path` and `operator.write` to render the topic,
// and the publish to the rendered topic.
// It is worth noting that in order to reduce publishing,
// only when the value of the writable property changes will be pushed.
func (d *mqttDevice) refreshAsAttributedTopic(newSpec v1alpha1.MQTTDeviceSpec) error {
	var newStatus v1alpha1.MQTTDeviceStatus

	var staleStatus = d.instance.Status
	var staleSpec = d.instance.Spec
	if !reflect.DeepEqual(staleSpec.Properties, newSpec.Properties) {
		d.stopFetch()

		var staleSpecPropsMap = mapSpecProperties(staleSpec.Properties)
		var staleStatusPropsMap = mapStatusProperties(staleStatus.Properties)

		// syncs properties
		var statusProps = make([]v1alpha1.MQTTDeviceStatusProperty, 0, len(newSpec.Properties))
		for i := 0; i < len(newSpec.Properties); i++ {
			var specPropPtr = &newSpec.Properties[i]
			var statusProp v1alpha1.MQTTDeviceStatusProperty
			if staleStatusPropPtr, existed := staleStatusPropsMap[specPropPtr.Name]; existed {
				statusProp = *staleStatusPropPtr
			} else {
				statusProp = v1alpha1.MQTTDeviceStatusProperty{
					Name:        specPropPtr.Name,
					Type:        specPropPtr.Type,
					AccessModes: specPropPtr.AccessModes,
					UpdatedAt:   now(),
				}
			}

			for _, accessMode := range specPropPtr.MergeAccessModes() {
				switch accessMode {
				case v1alpha1.MQTTDevicePropertyAccessModeWriteOnce:
					if !reflect.DeepEqual(specPropPtr, staleSpecPropsMap[specPropPtr.Name]) {
						var err = d.publishAttributedTopicProperty(specPropPtr)
						if err != nil {
							return errors.Wrapf(err, "failed to write property %s", specPropPtr.Name)
						}
					}
				case v1alpha1.MQTTDevicePropertyAccessModeWriteMany:
					var err = d.publishAttributedTopicProperty(specPropPtr)
					if err != nil {
						return errors.Wrapf(err, "failed to write property %s", specPropPtr.Name)
					}
				case v1alpha1.MQTTDevicePropertyAccessModeReadOnce:
					if !reflect.DeepEqual(specPropPtr, staleSpecPropsMap[specPropPtr.Name]) {
						var err = d.subscribeAttributedTopicProperty(specPropPtr, i, true)
						if err != nil {
							return errors.Wrapf(err, "failed to read property %s", specPropPtr.Name)
						}
					}
				default: // MQTTDevicePropertyAccessModeNotify
					var err = d.subscribeAttributedTopicProperty(specPropPtr, i, false)
					if err != nil {
						return errors.Wrapf(err, "failed to subscribe property %s", specPropPtr.Name)
					}
				}
			}

			statusProps = append(statusProps, statusProp)
		}
		newStatus = v1alpha1.MQTTDeviceStatus{Properties: statusProps}
	} else {
		newStatus = staleStatus
	}

	// fetches in backend
	if err := d.startFetch(newSpec.Protocol.GetSyncInterval()); err != nil {
		return errors.Wrap(err, "failed to start fetch")
	}

	// records
	d.instance.Spec = newSpec
	d.instance.Status = newStatus
	return d.sync()
}

// fetch is blocked, it is used to sync the MQTT device status periodically,
// it's worth noting that it just reads or writes the "WriteMany" properties.
func (d *mqttDevice) fetch(interval time.Duration, stop <-chan struct{}) {
	defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

	d.log.Info("Fetching")
	defer func() {
		d.log.Info("Finished fetching")
	}()

	var ticker = time.NewTicker(interval)
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

			// NB(thxCode) when the `spec.protocol` changes,
			// the `spec.properties` will be reset,
			// after obtaining the lock, this `fetch` goroutine should end.
			if len(d.instance.Status.Properties) != len(d.instance.Spec.Properties) {
				return
			}

			switch d.instance.Spec.Protocol.Pattern {
			case v1alpha1.MQTTDevicePatternAttributedMessage:
				var writablePayload []byte
				for i := 0; i < len(d.instance.Status.Properties); i++ {
					var specPropPtr = &d.instance.Spec.Properties[i]
					var err error
					for _, accessMode := range specPropPtr.MergeAccessModes() {
						if accessMode == v1alpha1.MQTTDevicePropertyAccessModeWriteMany {
							writablePayload, err = d.constructAttributedMessagePayload(writablePayload, specPropPtr)
							if err != nil {
								// TODO give a way to feedback this to limb.
								d.log.Error(err, "Error writing property", "property", specPropPtr.Name)
							}
						}
						// NB(thxCode) MergeAccessModes has already processed the accessModes,
						// and in this loop, we only need to process the "WriteMany" property.
						break
					}
				}
				if len(writablePayload) != 0 {
					if err := d.mqttClient.Publish(mqtt.PublishMessage{Payload: writablePayload}); err != nil {
						// TODO give a way to feedback this to limb.
						d.log.Error(err, "Error writing properties", "payload", string(writablePayload))
						return
					}
					d.log.V(4).Info("Write properties", "payload", string(writablePayload))
				}
			case v1alpha1.MQTTDevicePatternAttributeTopic:
				for i := 0; i < len(d.instance.Status.Properties); i++ {
					var specPropPtr = &d.instance.Spec.Properties[i]
					for _, accessMode := range specPropPtr.MergeAccessModes() {
						if accessMode == v1alpha1.MQTTDevicePropertyAccessModeWriteMany {
							var err = d.publishAttributedTopicProperty(specPropPtr)
							if err != nil {
								// TODO give a way to feedback this to limb.
								d.log.Error(err, "Error writing property", "property", specPropPtr.Name)
							}
						}
						// NB(thxCode) MergeAccessModes has already processed the accessModes,
						// and in this loop, we only need to process the "WriteMany" property.
						break
					}
				}
			default:
				return
			}

			if err := d.sync(); err != nil {
				d.log.Error(err, "Failed to sync")
			}
		}()

		select {
		case <-d.stop:
			return
		default:
		}
	}
}

func (d *mqttDevice) constructAttributedMessagePayload(payload []byte, propPtr *v1alpha1.MQTTDeviceProperty) ([]byte, error) {
	if propPtr.Value != nil {
		var propPath = getPath(propPtr.Name, propPtr.Visitor.Path)
		if err := verifyWritableJSONPath(propPath); err != nil {
			return nil, errors.Wrapf(err, "illegal JSON path %s", propPath)
		}
		var err error
		payload, err = sjson.SetBytes(payload, propPath, propPtr.Value)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to set value into path %s", propPath)
		}
		d.log.V(4).Info("Construct property into payload", "property", propPtr.Name, "type", propPtr.Type, "value", propPtr.Value)
	}

	return payload, nil
}

// publishAttributedTopicProperty writes data of a property to device.
func (d *mqttDevice) publishAttributedTopicProperty(propPtr *v1alpha1.MQTTDeviceProperty) error {
	if propPtr.Value != nil {
		var data, err = convertBytesContentTypeValueToBytes(propPtr)
		if err != nil {
			return err
		}

		err = d.mqttClient.Publish(mqtt.PublishMessage{
			Render:          getPublishRender(propPtr),
			QoSPointer:      propPtr.Visitor.GetQoSPtr(),
			RetainedPointer: propPtr.Visitor.Retained,
			Payload:         data,
		})
		if err != nil {
			return err
		}
		d.log.V(4).Info("Write property", "property", propPtr.Name, "type", propPtr.Type, "value", propPtr.Value)
	}

	return nil
}

// subscribeAttributedTopicProperty subscribes a property to receive the changes from device.
func (d *mqttDevice) subscribeAttributedTopicProperty(propPtr *v1alpha1.MQTTDeviceProperty, index int, once bool) error {
	return d.mqttClient.Subscribe(mqtt.SubscribeTopic{
		Index:      index,
		Render:     getSubscribeRender(propPtr),
		QoSPointer: propPtr.Visitor.GetQoSPtr(),
		Once:       once,
	})
}

// stopFetch stops the asynchronous fetch.
func (d *mqttDevice) stopFetch() {
	if d.stop == nil {
		return
	}

	// closes fetching
	close(d.stop)
	d.stop = nil

	// unsubscribe all subscriptions
	if d.mqttClient != nil {
		if err := d.mqttClient.StopSubscriptions(); err != nil {
			d.log.Error(err, "Failed to unsubscribe all subscriptions")
		}
	}
}

// startFetch starts the asynchronous fetch.
func (d *mqttDevice) startFetch(interval time.Duration) error {
	if d.stop != nil {
		return nil
	}
	d.stop = make(chan struct{})

	// starts fetching
	go d.fetch(interval, d.stop)

	// starts all subscriptions
	if d.mqttClient != nil {
		var subscribeHandler = func(msg mqtt.SubscribeMessage) {
			defer runtime.HandleCrash(handler.NewPanicsCleanupSocketHandler(metadata.Endpoint))

			// receives and updates status properties
			d.Lock()
			defer d.Unlock()

			if len(d.instance.Status.Properties) != len(d.instance.Spec.Properties) ||
				msg.Index >= len(d.instance.Status.Properties) {
				return
			}

			switch d.instance.Spec.Protocol.Pattern {
			case v1alpha1.MQTTDevicePatternAttributedMessage:
				for i := 0; i < len(d.instance.Status.Properties); i++ {
					var (
						specPropPtr   = &d.instance.Spec.Properties[i]
						statusPropPtr = &d.instance.Status.Properties[i]
					)
					for _, accessMode := range specPropPtr.MergeAccessModes() {
						switch accessMode {
						case v1alpha1.MQTTDevicePropertyAccessModeReadOnce:
							if statusPropPtr.Value != "" {
								break
							}
							fallthrough
						case v1alpha1.MQTTDevicePropertyAccessModeNotify:
							var propPath = getPath(specPropPtr.Name, specPropPtr.Visitor.Path)
							var propValueJSON = gjson.GetBytes(msg.Payload, propPath)
							var propValuePayload []byte
							if propValueJSON.Index > 0 {
								propValuePayload = msg.Payload[propValueJSON.Index : propValueJSON.Index+len(propValueJSON.Raw)]
							} else {
								propValuePayload = converter.UnsafeStringToBytes(propValueJSON.Raw)
							}

							var value, operationResult, err = parseTextContentTypeValueFromBytes(propValuePayload, specPropPtr)
							if err != nil {
								// TODO give a way to feedback this to limb.
								d.log.Error(err, "Error converting the byte array to property value")
								break
							}
							if accessMode == v1alpha1.MQTTDevicePropertyAccessModeNotify {
								d.log.V(4).Info("Notify property", "property", specPropPtr.Name, "type", specPropPtr.Type, "value", value, "operationResult", operationResult)
							} else {
								d.log.V(4).Info("Read property", "property", specPropPtr.Name, "type", specPropPtr.Type, "value", value, "operationResult", operationResult)
							}

							d.instance.Status.Properties[i] = v1alpha1.MQTTDeviceStatusProperty{
								Name:            specPropPtr.Name,
								Type:            specPropPtr.Type,
								AccessModes:     specPropPtr.AccessModes,
								Value:           value,
								OperationResult: operationResult,
								UpdatedAt:       now(),
							}
						}
					}
				}
			case v1alpha1.MQTTDevicePatternAttributeTopic:
				var specPropPtr = &d.instance.Spec.Properties[msg.Index]

				var (
					value           string
					operationResult string
					err             error
				)
				if specPropPtr.Visitor.ContentType == v1alpha1.MQTTDevicePropertyValueContentTypeBytes {
					value, operationResult, err = parseBytesContentTypeValueFromBytes(msg.Payload, specPropPtr)
				} else {
					value, operationResult, err = parseTextContentTypeValueFromBytes(msg.Payload, specPropPtr)
				}
				if err != nil {
					// TODO give a way to feedback this to limb.
					d.log.Error(err, "Error converting the byte array to property value")
					return
				}
				d.log.V(4).Info("Notify property", "property", specPropPtr.Name, "type", specPropPtr.Type, "value", value, "operationResult", operationResult)

				d.instance.Status.Properties[msg.Index] = v1alpha1.MQTTDeviceStatusProperty{
					Name:            specPropPtr.Name,
					Type:            specPropPtr.Type,
					AccessModes:     specPropPtr.AccessModes,
					Value:           value,
					OperationResult: operationResult,
					UpdatedAt:       now(),
				}
			default:
				return
			}

			// TODO we need to debounce here
			if err := d.sync(); err != nil {
				d.log.Error(err, "failed to sync")
			}
		}

		if err := d.mqttClient.StartSubscriptions(subscribeHandler); err != nil {
			return err
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
	render["path"] = getPath(prop.Name, prop.Visitor.Path)

	// gets operator rendering value
	if prop.Visitor.Operator != nil {
		var write = prop.Visitor.Operator.Write
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
	render["path"] = getPath(prop.Name, prop.Visitor.Path)

	// gets operator rendering value
	if prop.Visitor.Operator != nil {
		var read = prop.Visitor.Operator.Read
		if read == "null" {
			read = ""
		}
		render["operator"] = read
	}

	return render
}

func mapSpecProperties(specProps []v1alpha1.MQTTDeviceProperty) map[string]*v1alpha1.MQTTDeviceProperty {
	var ret = make(map[string]*v1alpha1.MQTTDeviceProperty, len(specProps))
	for i := 0; i < len(specProps); i++ {
		var prop = specProps[i]
		ret[prop.Name] = &prop
	}
	return ret
}

func mapStatusProperties(statusProps []v1alpha1.MQTTDeviceStatusProperty) map[string]*v1alpha1.MQTTDeviceStatusProperty {
	var ret = make(map[string]*v1alpha1.MQTTDeviceStatusProperty, len(statusProps))
	for i := 0; i < len(statusProps); i++ {
		var prop = statusProps[i]
		ret[prop.Name] = &prop
	}
	return ret
}

func now() *metav1.Time {
	var ret = metav1.Now()
	return &ret
}

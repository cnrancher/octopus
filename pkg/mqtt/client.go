package mqtt

import (
	"bytes"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	adaptorapi "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/mqtt/api"
	"github.com/rancher/octopus/pkg/util/converter"
)

// PublishMessage aggregates the parameters for publishing.
type PublishMessage struct {
	// Specifies the key-value pair for rendering topic.
	Render map[string]string

	// Specifies the QoS for publishing,
	// otherwise uses the global value.
	QoSPointer *byte

	// Specifies the retained for publishing,
	// otherwise uses the global value.
	RetainedPointer *bool

	// Specifies the payload.
	Payload interface{}
}

// SubscribeMessage aggregates the result from subscribing.
type SubscribeMessage struct {
	// Reports the index of corresponding SubscribeTopic.
	Index int

	// Reports the subscribed topic.
	Topic string

	// Reports the payload content.
	Payload []byte
}

// SubscribeHandler is a callback type which can be set to be
// executed upon the arrival of messages published to topics
// to which the client is subscribed.
type SubscribeHandler func(msg SubscribeMessage)

// SubscribeTopic aggregates the parameters for subscribing.
type SubscribeTopic struct {
	// Specifies the index, it can identify the message when subscribed arrived .
	Index int

	// Specifies the key-value pair for rendering topic.
	Render map[string]string

	// Specifies the QoS for publishing,
	// otherwise uses the global value.
	QoSPointer *byte

	// Specifies to subscribe the topic only once.
	Once bool
}

// SubscribeTopicIndex is an index of SubscribeTopic, which inspired by sets.String.
type SubscribeTopicIndex map[string]SubscribeTopic

// Insert adds topic to the index.
func (i SubscribeTopicIndex) Index(topicName string, topic *SubscribeTopic) {
	i[topicName] = *topic
}

// Difference returns a set of topic names that are not in o.
func (i SubscribeTopicIndex) DifferenceIndexes(o SubscribeTopicIndex) []string {
	var result = sets.String{}
	for topicName := range i {
		if _, contain := o[topicName]; !contain {
			result.Insert(topicName)
		}
	}
	return result.List()
}

// Len returns the amount of topics.
func (i SubscribeTopicIndex) Len() int {
	return len(i)
}

// PopAll pops the slice with topics in random order.
func (i SubscribeTopicIndex) PopAll() []string {
	var res = make([]string, 0, len(i))
	for key := range i {
		delete(i, key)
		res = append(res, key)
	}
	return res
}

type Client interface {
	// Connect will create a connection to the message broker.
	Connect() error

	// Disconnect will end the connection with the server.
	Disconnect()

	// RawClient returns the original MQTT client.
	RawClient() mqtt.Client

	// Publish publishes the message to corresponding topic.
	Publish(message PublishMessage) error

	// Subscribe subscribes the corresponding topic.
	Subscribe(topics ...SubscribeTopic) error

	// StartSubscriptions starts all subscriptions and handle in the same handler.
	StartSubscriptions(handler SubscribeHandler) error

	// StopSubscriptions stops all subscriptions.
	StopSubscriptions() error
}

type client struct {
	raw               mqtt.Client
	disconnectQuiesce time.Duration
	waitDuration      time.Duration
	topic             SegmentTopic
	qos               byte
	retained          bool

	subscribeTopicIndexer SubscribeTopicIndex
}

func (c *client) wait(token mqtt.Token) error {
	if c.waitDuration == 0 {
		_ = token.Wait()
	} else if !token.WaitTimeout(c.waitDuration) {
		return errors.Errorf("timeout in %v", c.waitDuration)
	}
	return token.Error()
}

func (c *client) Connect() error {
	var token = c.raw.Connect()
	_ = token.Wait()
	// NB(thxCode) we don't need to call token.WaitTimeout() in here as the connection timeout has been injected.
	return token.Error()
}

func (c *client) Disconnect() {
	var disconnectQuiesce = c.disconnectQuiesce
	if disconnectQuiesce == 0 {
		disconnectQuiesce = 5 * time.Second
	}
	c.raw.Disconnect(uint(disconnectQuiesce.Milliseconds()))
}

func (c *client) RawClient() mqtt.Client {
	return c.raw
}

func (c *client) Publish(message PublishMessage) error {
	if message.Payload == nil {
		return nil
	}
	var payload []byte
	switch p := message.Payload.(type) {
	case []byte:
		payload = p
	case bytes.Buffer:
		payload = p.Bytes()
	default:
		var encodedPayload, err = converter.MarshalJSON(message.Payload)
		if err != nil {
			return err
		}
		payload = encodedPayload
	}
	var qos = c.qos
	if message.QoSPointer != nil {
		qos = *message.QoSPointer
	}
	var retained = c.retained
	if message.RetainedPointer != nil {
		retained = *message.RetainedPointer
	}
	var topicName = c.topic.RenderForPublish(message.Render)
	log.Println("Publish  ", "topic: ", topicName, ", qos: ", qos, ", retained: ", retained)

	var token = c.raw.Publish(topicName, qos, retained, payload)
	return c.wait(token)
}

func (c *client) Subscribe(topics ...SubscribeTopic) error {
	for _, topic := range topics {
		var topicName = c.topic.RenderForSubscribe(topic.Render)
		c.subscribeTopicIndexer.Index(topicName, &topic)
	}
	return nil
}

func (c *client) StartSubscriptions(h SubscribeHandler) error {
	var topicIndexer = c.subscribeTopicIndexer
	if topicIndexer.Len() == 0 {
		return nil
	}

	// subscribes new topics
	var callback = func(cli mqtt.Client, msg mqtt.Message) {
		if h == nil {
			return
		}
		var topicName = msg.Topic()
		var topic, exist = topicIndexer[topicName]
		if !exist {
			return
		}

		log.Println("Receive message from topic: ", topicName)

		// unsubscribe if only subscribes once
		if topic.Once {
			go func() {
				var token = cli.Unsubscribe(topicName)
				_ = c.wait(token)
				log.Println("Unsubscribe topic: ", topicName)
			}()
		}

		// handles
		go h(SubscribeMessage{
			Index:   topic.Index,
			Topic:   topicName,
			Payload: msg.Payload(),
		})
	}

	var topicFilters = make(map[string]byte, len(topicIndexer))
	for topicName, topic := range topicIndexer {
		var qos = c.qos
		if topic.QoSPointer != nil {
			qos = *topic.QoSPointer
		}
		topicFilters[topicName] = qos
	}
	var token = c.raw.SubscribeMultiple(topicFilters, callback)
	if err := c.wait(token); err != nil {
		return err
	}

	log.Println("Subscribed topics: ", topicFilters)
	return nil
}

func (c *client) StopSubscriptions() error {
	var topicIndexer = c.subscribeTopicIndexer
	if topicIndexer.Len() == 0 {
		return nil
	}

	var topics = topicIndexer.PopAll()
	var token = c.raw.Unsubscribe(topics...)
	if err := c.wait(token); err != nil {
		return err
	}

	log.Println("Stop subscribed topics: ", topics)
	return nil
}

// NewClient creates the MQTT client with expected options.
func NewClient(spec api.MQTTOptions, ref corev1.ObjectReference, handler adaptorapi.ReferencesHandler) (Client, error) {
	var clientBuilder = NewClientBuilder(spec, ref)
	clientBuilder.Render(handler)
	return clientBuilder.Build()
}

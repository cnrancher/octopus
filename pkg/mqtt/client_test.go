package mqtt

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/rancher/octopus/pkg/mqtt/api"
	"github.com/rancher/octopus/pkg/mqtt/test"
	"github.com/rancher/octopus/pkg/util/log/zap"
)

func TestNewClient_Publish(t *testing.T) {
	type given struct {
		options api.MQTTOptions
		ref     corev1.ObjectReference
		payload interface{}
	}
	type expected struct {
		topic   string
		qos     byte
		payload []byte
	}

	var targetBrokerAddress = "tcp://127.0.0.1:8883"

	var testCases = []struct {
		name     string
		given    given
		expected expected
	}{
		{
			name: "payload's content-type is application/json",
			given: given{
				options: api.MQTTOptions{
					Client: api.MQTTClientOptions{
						Server: targetBrokerAddress,
					},
					Message: api.MQTTMessageOptions{
						Topic: "static/topic",
					},
				},
				ref: corev1.ObjectReference{
					Kind:       "DeviceLink",
					APIVersion: "edge.cattle.io/v1alpha1",
					Namespace:  "default",
					Name:       "test",
					UID:        "fcd1eb1b-ea42-4cb9-afb0-0ec2d0830583",
				},
				payload: map[string]string{
					"u": "u",
					"i": "i",
					"a": "a",
					"e": "e",
				},
			},
			expected: expected{
				topic:   "static/topic",
				qos:     1,
				payload: []byte(`{"a":"a","e":"e","i":"i","u":"u"}`),
			},
		},
		{
			name: "payload's content-type is default",
			given: given{
				options: api.MQTTOptions{
					Client: api.MQTTClientOptions{
						Server: targetBrokerAddress,
					},
					Message: api.MQTTMessageOptions{
						Topic: "dynamic/topic/:uid",
					},
				},
				ref: corev1.ObjectReference{
					Kind:       "DeviceLink",
					APIVersion: "edge.cattle.io/v1alpha1",
					Namespace:  "default",
					Name:       "test",
					UID:        "41478d1e-c3f8-46e3-a3b5-ba251f285277",
				},
				payload: "tested",
			},
			expected: expected{
				topic:   "dynamic/topic/41478d1e-c3f8-46e3-a3b5-ba251f285277",
				qos:     1,
				payload: []byte(`"tested"`),
			},
		},
	}

	var testBroker, err = test.NewMemoryBroker(targetBrokerAddress, zap.WrapAsLogr(zap.NewDevelopmentLogger()))
	assert.NoError(t, err)

	testBroker.Start()
	defer testBroker.Close()

	for _, tc := range testCases {
		var expectTopic = tc.expected.topic
		var expectQoS = tc.expected.qos
		var expectPayload = tc.expected.payload

		subscriptionStream, err := test.NewSubscriptionStream(targetBrokerAddress, expectTopic, expectQoS)
		assert.NoError(t, err, "case %q", tc.name)

		givenClient, err := NewClient(tc.given.options, tc.given.ref, nil)
		assert.NoError(t, err, "case %q", tc.name)

		err = givenClient.Connect()
		assert.NoError(t, err, "case %q", tc.name)

		err = givenClient.Publish(PublishMessage{
			Payload: tc.given.payload,
		})
		assert.NoError(t, err, "case %q, failed to publish payload", tc.name)

		err = subscriptionStream.Intercept(10*time.Second, func(actual *packet.Message) bool {
			return assert.Equal(t, expectPayload, actual.Payload, "case %q", tc.name)
		})
		subscriptionStream.Close()
		assert.NoError(t, err, "case %q", tc.name)
	}
}

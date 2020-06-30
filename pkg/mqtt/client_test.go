package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/mqtt/api/v1alpha1"
	"github.com/rancher/octopus/pkg/util/log/zap"
	"github.com/rancher/octopus/test/util/testdata"
)

type typeMetaObject struct {
	metav1.ObjectMeta
	metav1.TypeMeta
}

func Test_getTopic(t *testing.T) {
	type given struct {
		topic  v1alpha1.MQTTMessageTopic
		object typeMetaObject
	}
	type expected struct {
		ret string
		err error
	}

	var testCases = []struct {
		name     string
		given    given
		expected expected
	}{
		{
			name: "static name",
			given: given{
				topic: v1alpha1.MQTTMessageTopic{
					MQTTMessageTopicStatic: v1alpha1.MQTTMessageTopicStatic{
						Name: "static/topic",
					},
				},
				object: typeMetaObject{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "014997f5-1f12-498b-8631-d2f22920e20a",
					},
				},
			},
			expected: expected{
				ret: "static/topic",
			},
		},
		{
			name: "dynamic name with default mode",
			given: given{
				topic: v1alpha1.MQTTMessageTopic{
					MQTTMessageTopicDynamic: v1alpha1.MQTTMessageTopicDynamic{
						Prefix: "dynamic/topic",
					},
				},
				object: typeMetaObject{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "059d0d22-c78c-4dd4-9be7-f6aba119c234",
					},
				},
			},
			expected: expected{
				ret: "dynamic/topic/default/test",
			},
		},
		{
			name: "dynamic name with uid mode",
			given: given{
				topic: v1alpha1.MQTTMessageTopic{
					MQTTMessageTopicDynamic: v1alpha1.MQTTMessageTopicDynamic{
						Prefix: "dynamic/topic",
						With:   "uid",
					},
				},
				object: typeMetaObject{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "056a40a2-7cac-477e-99b1-8f06ecdce12a",
					},
				},
			},
			expected: expected{
				ret: "dynamic/topic/056a40a2-7cac-477e-99b1-8f06ecdce12a",
			},
		},
		{
			name: "both static and dynamic, but static first",
			given: given{
				topic: v1alpha1.MQTTMessageTopic{
					MQTTMessageTopicStatic: v1alpha1.MQTTMessageTopicStatic{
						Name: "static/topic",
					},
					MQTTMessageTopicDynamic: v1alpha1.MQTTMessageTopicDynamic{
						Prefix: "dynamic/topic",
					},
				},
				object: typeMetaObject{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "bb23d3dd-c36c-4b13-af8c-9ce8fb78dbb4",
					},
				},
			},
			expected: expected{
				ret: "static/topic",
			},
		},
		{
			name: "error if prefix is blank in dynamic mode",
			given: given{
				topic: v1alpha1.MQTTMessageTopic{
					MQTTMessageTopicDynamic: v1alpha1.MQTTMessageTopicDynamic{
						Prefix: "",
					},
				},
				object: typeMetaObject{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "9e8efcb0-4b79-45e7-a4c4-3349562292e3",
					},
				},
			},
			expected: expected{
				err: errors.New("topic prefix is required in dynamic mode"),
			},
		},
		{
			name: "error invalidate dynamic mode",
			given: given{
				topic: v1alpha1.MQTTMessageTopic{
					MQTTMessageTopicDynamic: v1alpha1.MQTTMessageTopicDynamic{
						Prefix: "dynamic/topic",
						With:   "unknown",
					},
				},
				object: typeMetaObject{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "fcd1eb1b-ea42-4cb9-afb0-0ec2d0830583",
					},
				},
			},
			expected: expected{
				err: errors.New("invalidate dynamic mode: unknown"),
			},
		},
	}

	for _, tc := range testCases {
		var actual, actualErr = getTopic(&tc.given.object, tc.given.topic)
		assert.Equal(t, tc.expected.ret, actual, "case %q", tc.name)
		assert.Equal(t, fmt.Sprint(tc.expected.err), fmt.Sprint(actualErr), "case %q", tc.name)
	}
}

func Test_getClientOptions(t *testing.T) {
	type given struct {
		uid        string
		topic      string
		options    v1alpha1.MQTTClientOptions
		references map[string]map[string][]byte
	}
	type expected struct {
		ret *mqtt.ClientOptions
		err error
	}

	var testCases = []struct {
		name     string
		given    given
		expected expected
	}{
		{
			name: "will message with static topic name",
			given: given{
				options: v1alpha1.MQTTClientOptions{
					Will: &v1alpha1.MQTTClientWillMessage{
						Topic: &v1alpha1.MQTTMessageTopicStatic{
							Name: "static/topic-will",
						},
						PayloadContent: "Y2xvc2Vk",
						PayloadEncode:  "base64",
						QoS:            1,
						Retained:       pointer.BoolPtr(true),
					},
				},
				uid:   "835aea2e-5f80-4d14-88f5-40c4bda41aa3",
				topic: "static/topic",
			},
			expected: expected{
				ret: mqtt.NewClientOptions().
					SetClientID("octopus-835aea2e5f804d1488f540c4bda41aa3").
					SetBinaryWill("static/topic-will", []byte("closed"), 1, true),
			},
		},
		{
			name: "will message without static topic name",
			given: given{
				options: v1alpha1.MQTTClientOptions{
					Will: &v1alpha1.MQTTClientWillMessage{
						PayloadContent: "closed",
						QoS:            1,
					},
				},
				uid:   "014997f5-1f12-498b-8631-d2f22920e20a",
				topic: "static/topic",
			},
			expected: expected{
				ret: mqtt.NewClientOptions().
					SetClientID("octopus-014997f51f12498b8631d2f22920e20a").
					SetBinaryWill("static/topic/$will", []byte("closed"), 1, false),
			},
		},
		{
			name: "configures HTTP header",
			given: given{
				options: v1alpha1.MQTTClientOptions{
					HTTPHeaders: map[string][]string{
						"header1": {"value1"},
					},
				},
				uid:   "bb23d3dd-c36c-4b13-af8c-9ce8fb78dbb4",
				topic: "static/topic",
			},
			expected: expected{
				ret: mqtt.NewClientOptions().
					SetClientID("octopus-bb23d3ddc36c4b13af8c9ce8fb78dbb4").
					SetHTTPHeaders(map[string][]string{
						"header1": {"value1"},
					}),
			},
		},
		{
			name: "configures TLS",
			given: given{
				options: v1alpha1.MQTTClientOptions{
					TLSConfig: &v1alpha1.MQTTClientTLS{
						CAFilePEM:   testdata.MustLoadString("ca.pem", t),
						CertFilePEM: testdata.MustLoadString("client.pem", t),
						KeyFilePEM:  testdata.MustLoadString("client-key.pem", t),
					},
				},
				uid:   "41478d1e-c3f8-46e3-a3b5-ba251f285277",
				topic: "static/topic",
			},
			expected: expected{
				ret: mqtt.NewClientOptions().
					SetClientID("octopus-41478d1ec3f846e3a3b5ba251f285277").
					SetTLSConfig(func() *tls.Config {
						var caPool = x509.NewCertPool()
						_ = caPool.AppendCertsFromPEM(testdata.MustLoadBytes("ca.pem", t))

						var cert, _ = tls.X509KeyPair(
							testdata.MustLoadBytes("client.pem", t),
							testdata.MustLoadBytes("client-key.pem", t),
						)

						return &tls.Config{
							RootCAs:      caPool,
							Certificates: []tls.Certificate{cert},
						}
					}()),
			},
		},
		{
			name: "configures TLS via references",
			given: given{
				options: v1alpha1.MQTTClientOptions{
					TLSConfig: &v1alpha1.MQTTClientTLS{
						CAFilePEMRef: &edgev1alpha1.DeviceLinkReferenceRelationship{
							Name: "ca",
							Item: "pem",
						},
						CertFilePEMRef: &edgev1alpha1.DeviceLinkReferenceRelationship{
							Name: "tls",
							Item: "cert",
						},
						KeyFilePEMRef: &edgev1alpha1.DeviceLinkReferenceRelationship{
							Name: "tls",
							Item: "key",
						},
					},
				},
				uid:   "41478d1e-c3f8-46e3-a3b5-ba251f285277",
				topic: "static/topic",
				references: map[string]map[string][]byte{
					"ca": {
						"pem": testdata.MustLoadBytes("ca.pem", t),
					},
					"tls": {
						"cert": testdata.MustLoadBytes("client.pem", t),
						"key":  testdata.MustLoadBytes("client-key.pem", t),
					},
				},
			},
			expected: expected{
				ret: mqtt.NewClientOptions().
					SetClientID("octopus-41478d1ec3f846e3a3b5ba251f285277").
					SetTLSConfig(func() *tls.Config {
						var caPool = x509.NewCertPool()
						_ = caPool.AppendCertsFromPEM(testdata.MustLoadBytes("ca.pem", t))

						var cert, _ = tls.X509KeyPair(
							testdata.MustLoadBytes("client.pem", t),
							testdata.MustLoadBytes("client-key.pem", t),
						)

						return &tls.Config{
							RootCAs:      caPool,
							Certificates: []tls.Certificate{cert},
						}
					}()),
			},
		},
		{
			name: "configures storage type as file",
			given: given{
				options: v1alpha1.MQTTClientOptions{
					Store: &v1alpha1.MQTTClientStore{
						Type:            "file",
						DirectoryPrefix: "/tmp/test",
					},
				},
				uid:   "056a40a2-7cac-477e-99b1-8f06ecdce12a",
				topic: "static/topic",
			},
			expected: expected{
				ret: mqtt.NewClientOptions().
					SetClientID("octopus-056a40a27cac477e99b18f06ecdce12a").
					SetStore(mqtt.NewFileStore("/tmp/test/056a40a2-7cac-477e-99b1-8f06ecdce12a")),
			},
		},
		{
			name: "configures forced client ID",
			given: given{
				options: v1alpha1.MQTTClientOptions{
					ClientID: "octopus-fake-clientid",
				},
				uid:   "2fd0fbd5-ba11-4f6c-aea8-b5c2c03b05c1",
				topic: "static/topic",
			},
			expected: expected{
				ret: mqtt.NewClientOptions().
					SetClientID("octopus-fake-clientid"),
			},
		},
		{
			name: "configures protocol version as 3",
			given: given{
				options: v1alpha1.MQTTClientOptions{
					ProtocolVersion: func() *uint { var i uint = 3; return &i }(),
				},
				uid:   "9e8efcb0-4b79-45e7-a4c4-3349562292e3",
				topic: "static/topic",
			},
			expected: expected{
				ret: mqtt.NewClientOptions().
					SetProtocolVersion(3).
					SetClientID("octopus-efcb04957c34593"),
			},
		},
	}

	// NB(thxCode) according to https://golang.org/src/reflect/deepequal.go, the Func values are equal only if both are nil.
	// We only focus on attribute comparison, so we can set the Func values as nil.
	var cleanFuncs = func(options *mqtt.ClientOptions) *mqtt.ClientOptions {
		if options != nil {
			options.SetCredentialsProvider(nil)
			options.SetDefaultPublishHandler(nil)
			options.SetOnConnectHandler(nil)
			options.SetConnectionLostHandler(nil)
		}
		return options
	}

	for _, tc := range testCases {
		var _, actual, actualErr = getClientOptionsOutline(tc.given.uid, tc.given.topic, tc.given.options, tc.given.references)
		assert.Equal(t, cleanFuncs(tc.expected.ret), cleanFuncs(actual), "case %q", tc.name)
		assert.Equal(t, fmt.Sprint(tc.expected.err), fmt.Sprint(actualErr), "case %q", tc.name)
	}
}

func TestNewClient(t *testing.T) {
	type given struct {
		object  typeMetaObject
		payload interface{}
		options v1alpha1.MQTTOptionsSpec
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
			name: "payload is in form of JSON",
			given: given{
				object: typeMetaObject{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "fcd1eb1b-ea42-4cb9-afb0-0ec2d0830583",
					},
				},
				payload: map[string]string{
					"u": "u",
					"i": "i",
					"a": "a",
					"e": "e",
				},
				options: v1alpha1.MQTTOptionsSpec{
					Client: v1alpha1.MQTTClientOptions{
						Server: targetBrokerAddress,
					},
					Message: v1alpha1.MQTTMessageOptions{
						Topic: v1alpha1.MQTTMessageTopic{
							MQTTMessageTopicStatic: v1alpha1.MQTTMessageTopicStatic{
								Name: "static/topic",
							},
						},
					},
				},
			},
			expected: expected{
				topic:   "static/topic",
				qos:     1,
				payload: []byte(`{"a":"a","e":"e","i":"i","u":"u"}`),
			},
		},
		{
			name: "payload is encoded by Base64",
			given: given{
				object: typeMetaObject{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "41478d1e-c3f8-46e3-a3b5-ba251f285277",
					},
				},
				payload: "tested",
				options: v1alpha1.MQTTOptionsSpec{
					Client: v1alpha1.MQTTClientOptions{
						Server: targetBrokerAddress,
					},
					Message: v1alpha1.MQTTMessageOptions{
						Topic: v1alpha1.MQTTMessageTopic{
							MQTTMessageTopicDynamic: v1alpha1.MQTTMessageTopicDynamic{
								Prefix: "dynamic/topic",
								With:   "uid",
							},
						},
						MQTTMessagePayloadOptions: v1alpha1.MQTTMessagePayloadOptions{
							PayloadEncode: "base64",
						},
					},
				},
			},
			expected: expected{
				topic:   "dynamic/topic/41478d1e-c3f8-46e3-a3b5-ba251f285277",
				qos:     1,
				payload: []byte("dGVzdGVk"),
			},
		},
	}

	var testBroker, err = NewTestMemoryBroker(targetBrokerAddress, zap.WrapAsLogr(zap.NewDevelopmentLogger()))
	assert.NoError(t, err)

	testBroker.Start()
	defer testBroker.Close()

	for _, tc := range testCases {
		var expectTopic = tc.expected.topic
		var expectQoS = tc.expected.qos
		var expectPayload = tc.expected.payload

		subscriptionStream, err := NewTestSubscriptionStream(targetBrokerAddress, expectTopic, expectQoS)
		assert.NoError(t, err, "case %q", tc.name)

		givenClient, _, err := NewClient(&tc.given.object, tc.given.options, nil)
		assert.NoError(t, err, "case %q", tc.name)

		err = givenClient.Connect()
		assert.NoError(t, err, "case %q", tc.name)

		if tc.given.payload != nil {
			err = givenClient.Publish(tc.given.payload)
			assert.NoError(t, err, "case %q, failed to publish payload", tc.name)
		} else {
			givenClient.Disconnect(1 * time.Second)
		}

		err = subscriptionStream.Intercept(10*time.Second, func(actual *packet.Message) bool {
			return assert.Equal(t, expectPayload, actual.Payload, "case %q", tc.name)
		})
		subscriptionStream.Close()
		assert.NoError(t, err, "case %q", tc.name)
	}
}

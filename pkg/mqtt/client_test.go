package mqtt

import (
	"crypto/tls"
	"crypto/x509"
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

type testMetaObj struct {
	metav1.ObjectMeta
	metav1.TypeMeta
}

func Test_getTopic(t *testing.T) {
	type given struct {
		topic  v1alpha1.MQTTMessageTopic
		object testMetaObj
	}

	type expect struct {
		topic string
		err   error
	}

	var testCases = []struct {
		given  given
		expect expect
	}{
		{ // static
			given: given{
				topic: v1alpha1.MQTTMessageTopic{
					MQTTMessageTopicStatic: v1alpha1.MQTTMessageTopicStatic{
						Name: "static/topic",
					},
				},
				object: testMetaObj{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "014997f5-1f12-498b-8631-d2f22920e20a",
					},
				},
			},
			expect: expect{
				topic: "static/topic",
			},
		},
		{ // dynamic with default mode
			given: given{
				topic: v1alpha1.MQTTMessageTopic{
					MQTTMessageTopicDynamic: v1alpha1.MQTTMessageTopicDynamic{
						Prefix: "dynamic/topic",
					},
				},
				object: testMetaObj{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "059d0d22-c78c-4dd4-9be7-f6aba119c234",
					},
				},
			},
			expect: expect{
				topic: "dynamic/topic/default/test",
			},
		},
		{ // dynamic with uid mode
			given: given{
				topic: v1alpha1.MQTTMessageTopic{
					MQTTMessageTopicDynamic: v1alpha1.MQTTMessageTopicDynamic{
						Prefix: "dynamic/topic",
						With:   "uid",
					},
				},
				object: testMetaObj{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "056a40a2-7cac-477e-99b1-8f06ecdce12a",
					},
				},
			},
			expect: expect{
				topic: "dynamic/topic/056a40a2-7cac-477e-99b1-8f06ecdce12a",
			},
		},
		{ // static first
			given: given{
				topic: v1alpha1.MQTTMessageTopic{
					MQTTMessageTopicStatic: v1alpha1.MQTTMessageTopicStatic{
						Name: "static/topic",
					},
					MQTTMessageTopicDynamic: v1alpha1.MQTTMessageTopicDynamic{
						Prefix: "dynamic/topic",
					},
				},
				object: testMetaObj{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "bb23d3dd-c36c-4b13-af8c-9ce8fb78dbb4",
					},
				},
			},
			expect: expect{
				topic: "static/topic",
			},
		},
		{ // error if prefix is blank in dynamic mode
			given: given{
				topic: v1alpha1.MQTTMessageTopic{
					MQTTMessageTopicDynamic: v1alpha1.MQTTMessageTopicDynamic{
						Prefix: "",
					},
				},
				object: testMetaObj{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "9e8efcb0-4b79-45e7-a4c4-3349562292e3",
					},
				},
			},
			expect: expect{
				err: errors.New("topic prefix is required in dynamic mode"),
			},
		},
		{ // error invalidate dynamic mode
			given: given{
				topic: v1alpha1.MQTTMessageTopic{
					MQTTMessageTopicDynamic: v1alpha1.MQTTMessageTopicDynamic{
						Prefix: "dynamic/topic",
						With:   "unknown",
					},
				},
				object: testMetaObj{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "fcd1eb1b-ea42-4cb9-afb0-0ec2d0830583",
					},
				},
			},
			expect: expect{
				err: errors.New("invalidate dynamic mode: unknown"),
			},
		},
	}

	for i, tc := range testCases {
		var actualTopic, actualErr = getTopic(&tc.given.object, tc.given.topic)
		if tc.expect.err != nil || actualErr != nil {
			assert.EqualError(t, actualErr, tc.expect.err.Error(), "case %v, failed to get topic", i+1)
			continue
		}
		assert.Equal(t, tc.expect.topic, actualTopic, "case %v, failed to compare with two topics", i+1)
	}
}

func Test_getClientOptions(t *testing.T) {
	type given struct {
		uid        string
		topic      string
		options    v1alpha1.MQTTClientOptions
		references map[string]map[string][]byte
	}

	type expect struct {
		options *mqtt.ClientOptions
		err     error
	}

	var testCases = []struct {
		given  given
		expect expect
	}{
		{ // will with topic
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
			expect: expect{
				options: mqtt.NewClientOptions().
					SetClientID("octopus-835aea2e5f804d1488f540c4bda41aa3").
					SetBinaryWill("static/topic-will", []byte("closed"), 1, true),
			},
		},
		{ // will without topic
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
			expect: expect{
				options: mqtt.NewClientOptions().
					SetClientID("octopus-014997f51f12498b8631d2f22920e20a").
					SetBinaryWill("static/topic/$will", []byte("closed"), 1, false),
			},
		},
		{ // header
			given: given{
				options: v1alpha1.MQTTClientOptions{
					HTTPHeaders: map[string][]string{
						"header1": {"value1"},
					},
				},
				uid:   "bb23d3dd-c36c-4b13-af8c-9ce8fb78dbb4",
				topic: "static/topic",
			},
			expect: expect{
				options: mqtt.NewClientOptions().
					SetClientID("octopus-bb23d3ddc36c4b13af8c9ce8fb78dbb4").
					SetHTTPHeaders(map[string][]string{
						"header1": {"value1"},
					}),
			},
		},
		{ // TLS
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
			expect: expect{
				options: mqtt.NewClientOptions().
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
		{ // TLS in references
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
			expect: expect{
				options: mqtt.NewClientOptions().
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
		{ // storage
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
			expect: expect{
				options: mqtt.NewClientOptions().
					SetClientID("octopus-056a40a27cac477e99b18f06ecdce12a").
					SetStore(mqtt.NewFileStore("/tmp/test/056a40a2-7cac-477e-99b1-8f06ecdce12a")),
			},
		},
		{ // clientID
			given: given{
				options: v1alpha1.MQTTClientOptions{
					ClientID: "octopus-fake-clientid",
				},
				uid:   "2fd0fbd5-ba11-4f6c-aea8-b5c2c03b05c1",
				topic: "static/topic",
			},
			expect: expect{
				options: mqtt.NewClientOptions().
					SetClientID("octopus-fake-clientid"),
			},
		},
		{ // clientID in MQTT v3.1
			given: given{
				options: v1alpha1.MQTTClientOptions{
					ProtocolVersion: func() *uint { var i uint = 3; return &i }(),
				},
				uid:   "9e8efcb0-4b79-45e7-a4c4-3349562292e3",
				topic: "static/topic",
			},
			expect: expect{
				options: mqtt.NewClientOptions().
					SetProtocolVersion(3).
					SetClientID("octopus-efcb04957c34593"),
			},
		},
	}

	for i, tc := range testCases {
		var _, actualOptions, actualErr = getClientOptionsOutline(tc.given.uid, tc.given.topic, tc.given.options, tc.given.references)
		if actualErr != nil {
			if tc.expect.err == nil {
				assert.NoError(t, actualErr, "case %v, failed to get client options", i+1)
			} else {
				assert.EqualError(t, actualErr, tc.expect.err.Error(), "case %v", i+1)
			}
			continue
		}

		// NB(thxCode) according to https://golang.org/src/reflect/deepequal.go, the Func values are equal only if both are nil.
		// We only focus on attribute comparison, so we can set the Func values as nil.
		var expectedOptions = tc.expect.options
		expectedOptions.SetCredentialsProvider(nil)
		actualOptions.SetCredentialsProvider(nil)
		expectedOptions.SetDefaultPublishHandler(nil)
		actualOptions.SetDefaultPublishHandler(nil)
		expectedOptions.SetOnConnectHandler(nil)
		actualOptions.SetOnConnectHandler(nil)
		expectedOptions.SetConnectionLostHandler(nil)
		actualOptions.SetConnectionLostHandler(nil)
		assert.Equal(t, expectedOptions, actualOptions, "case %v, failed to compare with two options", i+1)
	}
}

func TestNewClient(t *testing.T) {
	type given struct {
		object  testMetaObj
		payload interface{}
		options v1alpha1.MQTTOptionsSpec
	}

	type expect struct {
		topic   string
		qos     byte
		payload []byte
	}

	var testBrokerAddress = "tcp://127.0.0.1:8883"
	var testCases = []struct {
		given  given
		expect expect
	}{
		{ // json payload
			given: given{
				object: testMetaObj{
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
						Server: testBrokerAddress,
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
			expect: expect{
				topic:   "static/topic",
				qos:     1,
				payload: []byte(`{"a":"a","e":"e","i":"i","u":"u"}`),
			},
		},
		{ // payload encoded by base64
			given: given{
				object: testMetaObj{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "test",
						UID:       "41478d1e-c3f8-46e3-a3b5-ba251f285277",
					},
				},
				payload: "tested",
				options: v1alpha1.MQTTOptionsSpec{
					Client: v1alpha1.MQTTClientOptions{
						Server: testBrokerAddress,
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
			expect: expect{
				topic:   "dynamic/topic/41478d1e-c3f8-46e3-a3b5-ba251f285277",
				qos:     1,
				payload: []byte("dGVzdGVk"),
			},
		},
	}

	var testBroker, err = NewTestMemoryBroker(testBrokerAddress, zap.WrapAsLogr(zap.NewDevelopmentLogger()))
	assert.NoError(t, err)

	testBroker.Start()
	defer testBroker.Close()

	for i, tc := range testCases {
		var expectTopic = tc.expect.topic
		var expectQoS = tc.expect.qos
		var expectPayload = tc.expect.payload

		subscriptionStream, err := NewTestSubscriptionStream(testBrokerAddress, expectTopic, expectQoS)
		assert.NoError(t, err, "case %v", i+1)

		givenClient, _, err := NewClient(&tc.given.object, tc.given.options, nil)
		assert.NoError(t, err, "case %v", i+1)

		err = givenClient.Connect()
		assert.NoError(t, err, "case %v", i+1)

		if tc.given.payload != nil {
			err = givenClient.Publish(tc.given.payload)
			assert.NoError(t, err, "case %v, failed to publish payload", i+1)
		} else {
			givenClient.Disconnect(1 * time.Second)
		}

		err = subscriptionStream.Intercept(10*time.Second, func(actual *packet.Message) bool {
			return assert.Equal(t, expectPayload, actual.Payload, "case %v")
		})
		subscriptionStream.Close()
		assert.NoError(t, err, "case %v", i+1)
	}
}

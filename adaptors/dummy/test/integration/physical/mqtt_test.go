package adaptor

import (
	"fmt"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/rancher/octopus/adaptors/dummy/api/v1alpha1"
	"github.com/rancher/octopus/adaptors/dummy/pkg/physical"
	"github.com/rancher/octopus/pkg/mqtt"
	mqttapi "github.com/rancher/octopus/pkg/mqtt/api/v1alpha1"
	"github.com/rancher/octopus/pkg/util/converter"
	"github.com/rancher/octopus/pkg/util/log/zap"
	"github.com/rancher/octopus/pkg/util/object"
	"github.com/rancher/octopus/test/util/testdata"
)

// testing scenarios:
// + Special Device
// 		- validate if status changes can be subscribed
//      - validate if TLS configuration works
//      - validate if MQTT works when the configuration changes
// + Protocol Device
//		- validate if status changes can be subscribed
var _ = Describe("MQTT", func() {
	var (
		testUnencryptedBrokerAddress              = "tcp://test.mosquitto.org:1883"
		testEncryptedAndClientAuthedBrokerAddress = "tcps://test.mosquitto.org:8884"

		testNamespace = "default"

		err error
		log logr.Logger
	)

	JustBeforeEach(func() {
		log = zap.WrapAsLogr(zap.NewDevelopmentLogger())
	})

	Context("Special Device", func() {
		var (
			testInstance               *v1alpha1.DummySpecialDevice
			testInstanceNamespacedName types.NamespacedName
		)

		BeforeEach(func() {
			var timestamp = time.Now().Unix()
			testInstance = &v1alpha1.DummySpecialDevice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNamespace,
					Name:      fmt.Sprintf("s-%d", timestamp),
					UID:       types.UID(fmt.Sprintf("uid-%d", timestamp)),
				},
			}
			testInstanceNamespacedName = object.GetNamespacedName(testInstance)
		})

		It("should subscribe the status changes", func() {

			/*
				since the dummy special device can mocking the device' status change,
				we can just create an instance and keep watching if there is any subscribed message incomes.
			*/

			var testSubscriptionStream *mqtt.TestSubscriptionStream
			testSubscriptionStream, err = mqtt.NewTestSubscriptionStream(testUnencryptedBrokerAddress, fmt.Sprintf("cattle.io/octopus/%s", testInstanceNamespacedName), 0)
			Expect(err).ToNot(HaveOccurred())
			defer testSubscriptionStream.Close()

			var testDevice = physical.NewSpecialDevice(
				log.WithValues("device", testInstanceNamespacedName),
				testInstance,
				nil,
			)
			defer testDevice.Shutdown()

			var testSpec = v1alpha1.DummySpecialDeviceSpec{
				Extension: v1alpha1.DeviceExtensionSpec{
					MQTT: &mqttapi.MQTTOptionsSpec{
						Client: mqttapi.MQTTClientOptions{
							Server: testUnencryptedBrokerAddress,
						},
						Message: mqttapi.MQTTMessageOptions{
							// dynamic topic
							Topic: mqttapi.MQTTMessageTopic{
								MQTTMessageTopicDynamic: mqttapi.MQTTMessageTopicDynamic{
									Prefix: "cattle.io/octopus",
								},
							},
							MQTTMessagePayloadOptions: mqttapi.MQTTMessagePayloadOptions{
								QoS: 1,
							},
						},
					},
				},
				Protocol: v1alpha1.DummySpecialDeviceProtocol{
					Location: "127.0.0.1",
				},
				On:   true,
				Gear: v1alpha1.DummySpecialDeviceGearFast,
			}
			err = testDevice.Configure(nil, testSpec)
			Expect(err).ToNot(HaveOccurred())

			var receivedCount = 2
			err = testSubscriptionStream.Intercept(15*time.Second, func(actual *packet.Message) bool {
				GinkgoT().Logf("topic: %s, qos: %d, retain: %v, payload: %s", actual.Topic, actual.QOS, actual.Retain, converter.UnsafeBytesToString(actual.Payload))

				receivedCount--
				if receivedCount == 0 {
					return true
				}
				return false
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should connect via SSL/TLS", func() {

			/*
				since test.mosquitto.org provides a public CA, we can download it and generate the client cert key pair
				via https://test.mosquitto.org/ssl/index.php.
			*/

			var testSubscriptionStream *mqtt.TestSubscriptionStream
			// subscribe via unencrypted endpoint
			testSubscriptionStream, err = mqtt.NewTestSubscriptionStream(testUnencryptedBrokerAddress, fmt.Sprintf("cattle.io/octopus/%s", testInstance.GetUID()), 0)
			Expect(err).ToNot(HaveOccurred())
			defer testSubscriptionStream.Close()

			var testDevice = physical.NewSpecialDevice(
				log.WithValues("device", testInstanceNamespacedName),
				testInstance,
				nil,
			)
			defer testDevice.Shutdown()

			var testSpec = v1alpha1.DummySpecialDeviceSpec{
				Extension: v1alpha1.DeviceExtensionSpec{
					MQTT: &mqttapi.MQTTOptionsSpec{
						Client: mqttapi.MQTTClientOptions{
							// publish via encrypted endpoint
							Server: testEncryptedAndClientAuthedBrokerAddress,
							TLSConfig: &mqttapi.MQTTClientTLS{
								CAFilePEM:   testdata.MustLoadString("mosquitto.org.crt", GinkgoT()),
								CertFilePEM: testdata.MustLoadString("client.crt", GinkgoT()),
								KeyFilePEM:  testdata.MustLoadString("client-key.pem", GinkgoT()),
								ServerName:  "test.mosquitto.org",
							},
						},
						Message: mqttapi.MQTTMessageOptions{
							// dynamic topic with uid
							Topic: mqttapi.MQTTMessageTopic{
								MQTTMessageTopicDynamic: mqttapi.MQTTMessageTopicDynamic{
									Prefix: "cattle.io/octopus",
									With:   "uid",
								},
							},
						},
					},
				},
				Protocol: v1alpha1.DummySpecialDeviceProtocol{
					Location: "living-room",
				},
				On:   true,
				Gear: v1alpha1.DummySpecialDeviceGearFast,
			}
			err = testDevice.Configure(nil, testSpec)
			Expect(err).ToNot(HaveOccurred())

			err = testSubscriptionStream.Intercept(15*time.Second, func(actual *packet.Message) bool {
				GinkgoT().Logf("topic: %s, qos: %d, retain: %v, payload: %s", actual.Topic, actual.QOS, actual.Retain, converter.UnsafeBytesToString(actual.Payload))
				return true
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should work after changing settings", func() {

			/*
				we will use dynamic topic at first, and then change to static topic.
			*/

			/*
				dynamic topic with nn at first
			*/

			var testSubscriptionStream *mqtt.TestSubscriptionStream
			testSubscriptionStream, err = mqtt.NewTestSubscriptionStream(testUnencryptedBrokerAddress, fmt.Sprintf("cattle.io/octopus/%s", testInstanceNamespacedName), 0)
			Expect(err).ToNot(HaveOccurred())

			var testDevice = physical.NewSpecialDevice(
				log.WithValues("device", testInstanceNamespacedName),
				testInstance,
				nil,
			)
			defer testDevice.Shutdown()

			var testSpec = v1alpha1.DummySpecialDeviceSpec{
				Extension: v1alpha1.DeviceExtensionSpec{
					MQTT: &mqttapi.MQTTOptionsSpec{
						Client: mqttapi.MQTTClientOptions{
							Server: testUnencryptedBrokerAddress,
						},
						Message: mqttapi.MQTTMessageOptions{
							Topic: mqttapi.MQTTMessageTopic{
								MQTTMessageTopicDynamic: mqttapi.MQTTMessageTopicDynamic{
									Prefix: "cattle.io/octopus",
								},
							},
						},
					},
				},
				Protocol: v1alpha1.DummySpecialDeviceProtocol{
					Location: "living-room",
				},
				On:   true,
				Gear: v1alpha1.DummySpecialDeviceGearFast,
			}
			err = testDevice.Configure(nil, testSpec)
			Expect(err).ToNot(HaveOccurred())

			err = testSubscriptionStream.Intercept(15*time.Second, func(actual *packet.Message) bool {
				GinkgoT().Logf("topic: %s, qos: %d, retain: %v, payload: %s", actual.Topic, actual.QOS, actual.Retain, converter.UnsafeBytesToString(actual.Payload))
				return true
			})
			Expect(err).ToNot(HaveOccurred())
			testSubscriptionStream.Close()

			/*
				change to static topic
			*/

			testSubscriptionStream, err = mqtt.NewTestSubscriptionStream(testUnencryptedBrokerAddress, "cattle.io/octopus/default/test3/static", 0)
			Expect(err).ToNot(HaveOccurred())

			var testNewSpec = v1alpha1.DummySpecialDeviceSpec{
				Extension: v1alpha1.DeviceExtensionSpec{
					MQTT: &mqttapi.MQTTOptionsSpec{
						Client: mqttapi.MQTTClientOptions{
							Server: testUnencryptedBrokerAddress,
						},
						Message: mqttapi.MQTTMessageOptions{
							Topic: mqttapi.MQTTMessageTopic{
								MQTTMessageTopicStatic: mqttapi.MQTTMessageTopicStatic{
									Name: "cattle.io/octopus/default/test3/static",
								},
							},
						},
					},
				},
				Protocol: v1alpha1.DummySpecialDeviceProtocol{
					Location: "living-room",
				},
				On:   true,
				Gear: v1alpha1.DummySpecialDeviceGearMiddle,
			}
			err = testDevice.Configure(nil, testNewSpec)
			Expect(err).ToNot(HaveOccurred())

			err = testSubscriptionStream.Intercept(15*time.Second, func(actual *packet.Message) bool {
				GinkgoT().Logf("topic: %s, qos: %d, retain: %v, payload: %s", actual.Topic, actual.QOS, actual.Retain, converter.UnsafeBytesToString(actual.Payload))
				return true
			})
			Expect(err).ToNot(HaveOccurred())
			testSubscriptionStream.Close()
		})
	})

	Context("Protocol Device", func() {
		var (
			testInstance               *v1alpha1.DummyProtocolDevice
			testInstanceNamespacedName types.NamespacedName
		)

		BeforeEach(func() {
			var timestamp = time.Now().Unix()
			testInstance = &v1alpha1.DummyProtocolDevice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNamespace,
					Name:      fmt.Sprintf("p-%d", timestamp),
					UID:       types.UID(fmt.Sprintf("uid-%d", timestamp)),
				},
			}
			testInstanceNamespacedName = object.GetNamespacedName(testInstance)
		})

		It("should subscribe the status changes", func() {

			/*
				since the dummy protocol device can mocking the device' status change,
				we can just create an instance and keep watching if there is any subscribed message incomes.
			*/

			var testSubscriptionStream *mqtt.TestSubscriptionStream
			testSubscriptionStream, err = mqtt.NewTestSubscriptionStream(testUnencryptedBrokerAddress, fmt.Sprintf("cattle.io/octopus/%s", testInstanceNamespacedName), 0)
			Expect(err).ToNot(HaveOccurred())
			defer testSubscriptionStream.Close()

			var testDevice = physical.NewProtocolDevice(
				log.WithValues("device", testInstanceNamespacedName),
				testInstance,
				nil,
			)
			defer testDevice.Shutdown()

			var testSpec = v1alpha1.DummyProtocolDeviceSpec{
				Extension: v1alpha1.DeviceExtensionSpec{
					MQTT: &mqttapi.MQTTOptionsSpec{
						Client: mqttapi.MQTTClientOptions{
							Server: testUnencryptedBrokerAddress,
						},
						Message: mqttapi.MQTTMessageOptions{
							// dynamic topic
							Topic: mqttapi.MQTTMessageTopic{
								MQTTMessageTopicDynamic: mqttapi.MQTTMessageTopicDynamic{
									Prefix: "cattle.io/octopus",
								},
							},
							MQTTMessagePayloadOptions: mqttapi.MQTTMessagePayloadOptions{
								QoS: 1,
							},
						},
					},
				},
				Protocol: v1alpha1.DummyProtocolDeviceProtocol{
					IP: "192.168.3.6",
				},
				Props: map[string]v1alpha1.DummyProtocolDeviceSpecProps{
					"string": {
						Type: v1alpha1.DummyProtocolDevicePropertyTypeString,
					},
					"integer": {
						Type: v1alpha1.DummyProtocolDevicePropertyTypeInt,
					},
					"float": {
						Type: v1alpha1.DummyProtocolDevicePropertyTypeFloat,
					},
					"object": {
						Type: v1alpha1.DummyProtocolDevicePropertyTypeObject,
						ObjectProps: map[string]v1alpha1.DummyProtocolDeviceSpecObjectOrArrayProps{
							"objectString": {
								DummyProtocolDeviceSpecProps: v1alpha1.DummyProtocolDeviceSpecProps{
									Type: v1alpha1.DummyProtocolDevicePropertyTypeString,
								},
							},
							"objectInteger": {
								DummyProtocolDeviceSpecProps: v1alpha1.DummyProtocolDeviceSpecProps{
									Type: v1alpha1.DummyProtocolDevicePropertyTypeInt,
								},
							},
						},
					},
					"array": {
						Type: v1alpha1.DummyProtocolDevicePropertyTypeArray,
						ArrayProps: &v1alpha1.DummyProtocolDeviceSpecObjectOrArrayProps{
							DummyProtocolDeviceSpecProps: v1alpha1.DummyProtocolDeviceSpecProps{
								Type: v1alpha1.DummyProtocolDevicePropertyTypeInt,
							},
						},
					},
				},
			}
			err = testDevice.Configure(nil, testSpec)
			Expect(err).ToNot(HaveOccurred())

			var receivedCount = 2
			err = testSubscriptionStream.Intercept(15*time.Second, func(actual *packet.Message) bool {
				GinkgoT().Logf("topic: %s, qos: %d, retain: %v, payload: %s", actual.Topic, actual.QOS, actual.Retain, converter.UnsafeBytesToString(actual.Payload))

				receivedCount--
				if receivedCount == 0 {
					return true
				}
				return false
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

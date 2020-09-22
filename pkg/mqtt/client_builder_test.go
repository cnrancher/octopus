package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	adaptorapi "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/mqtt/api"
	"github.com/rancher/octopus/test/util/testdata"
)

func TestClientBuilder_Render(t *testing.T) {
	type given struct {
		spec    api.MQTTOptions
		ref     corev1.ObjectReference
		handler adaptorapi.ReferencesHandler
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
				spec: api.MQTTOptions{
					Message: api.MQTTMessageOptions{
						Will: &api.MQTTWillMessage{
							Topic: "static/topic-will",
							Content: api.MQTTWillMessageContent{
								Data: []byte("closed"),
							},
						},
					},
				},
				ref: corev1.ObjectReference{
					Namespace: "default",
					Name:      "test",
					UID:       "835aea2e-5f80-4d14-88f5-40c4bda41aa3",
				},
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
				spec: api.MQTTOptions{
					Message: api.MQTTMessageOptions{
						Topic: "static/topic",
						Will: &api.MQTTWillMessage{
							Content: api.MQTTWillMessageContent{
								Data: []byte("closed"),
							},
						},
					},
				},
				ref: corev1.ObjectReference{
					Namespace: "default",
					Name:      "test",
					UID:       "014997f5-1f12-498b-8631-d2f22920e20a",
				},
			},
			expected: expected{
				ret: mqtt.NewClientOptions().
					SetClientID("octopus-014997f51f12498b8631d2f22920e20a").
					SetBinaryWill("static/topic/$will", []byte("closed"), 1, true),
			},
		},
		{
			name: "configures HTTP header",
			given: given{
				spec: api.MQTTOptions{
					Client: api.MQTTClientOptions{
						HTTPHeaders: map[string][]string{
							"header1": {"value1"},
						},
					},
				},
				ref: corev1.ObjectReference{
					Namespace: "default",
					Name:      "test",
					UID:       "bb23d3dd-c36c-4b13-af8c-9ce8fb78dbb4",
				},
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
				spec: api.MQTTOptions{
					Client: api.MQTTClientOptions{
						TLSConfig: &api.MQTTClientTLS{
							CAFilePEM:   testdata.MustLoadString("ca.pem", t),
							CertFilePEM: testdata.MustLoadString("client.pem", t),
							KeyFilePEM:  testdata.MustLoadString("client-key.pem", t),
						},
					},
				},
				ref: corev1.ObjectReference{
					Namespace: "default",
					Name:      "test",
					UID:       "41478d1e-c3f8-46e3-a3b5-ba251f285277",
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
			name: "configures TLS via references",
			given: given{
				spec: api.MQTTOptions{
					Client: api.MQTTClientOptions{
						TLSConfig: &api.MQTTClientTLS{
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
				},
				ref: corev1.ObjectReference{
					Namespace: "default",
					Name:      "test",
					UID:       "41478d1e-c3f8-46e3-a3b5-ba251f285277",
				},
				handler: adaptorapi.ReferencesHandler(
					map[string]*adaptorapi.ConnectRequestReferenceEntry{
						"ca": {
							Items: map[string][]byte{
								"pem": testdata.MustLoadBytes("ca.pem", t),
							},
						},
						"tls": {
							Items: map[string][]byte{
								"cert": testdata.MustLoadBytes("client.pem", t),
								"key":  testdata.MustLoadBytes("client-key.pem", t),
							},
						},
					},
				),
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
				spec: api.MQTTOptions{
					Client: api.MQTTClientOptions{
						Store: &api.MQTTClientStore{
							Type:            "File",
							DirectoryPrefix: "/tmp/test",
						},
					},
				},
				ref: corev1.ObjectReference{
					Namespace: "default",
					Name:      "test",
					UID:       "056a40a2-7cac-477e-99b1-8f06ecdce12a",
				},
			},
			expected: expected{
				ret: mqtt.NewClientOptions().
					SetClientID("octopus-056a40a27cac477e99b18f06ecdce12a").
					SetStore(mqtt.NewFileStore("/tmp/test/056a40a2-7cac-477e-99b1-8f06ecdce12a")),
			},
		},
		{
			name: "configures protocol version as 3",
			given: given{
				spec: api.MQTTOptions{
					Client: api.MQTTClientOptions{
						ProtocolVersion: func() *uint { var i uint = 3; return &i }(),
					},
				},
				ref: corev1.ObjectReference{
					Namespace: "default",
					Name:      "test",
					UID:       "9e8efcb0-4b79-45e7-a4c4-3349562292e3",
				},
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
		var cb = NewClientBuilder(tc.given.spec, tc.given.ref)
		cb.Render(tc.given.handler)
		var _, err = cb.Build()
		if assert.Nil(t, err, "case %q", tc.name) {
			var actual = cb.GetOptions()
			assert.Equal(t, cleanFuncs(tc.expected.ret), cleanFuncs(actual), "case %q", tc.name)
		}
	}
}

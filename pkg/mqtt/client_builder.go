package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"path"
	"path/filepath"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	adaptorapi "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
	"github.com/rancher/octopus/pkg/mqtt/api"
	"github.com/rancher/octopus/pkg/util/converter"
	"github.com/rancher/octopus/pkg/util/uuid"
)

type CustomMQTTOptionFunc func(options *mqtt.ClientOptions) error

type ClientBuilder struct {
	ref    corev1.ObjectReference
	spec   *api.MQTTOptions
	status *mqtt.ClientOptions
	err    error
}

// Render renders the MQTT client options with expected options.
func (b *ClientBuilder) Render(handler adaptorapi.ReferencesHandler) {
	var ref = b.ref
	var clientSpec = b.spec.Client
	var messageSpec = b.spec.Message
	var status = b.status

	// processes basic authentication
	if clientSpec.BasicAuth != nil {
		var basicAuthSpec = clientSpec.BasicAuth

		var username string
		if basicAuthSpec.Username != "" {
			username = basicAuthSpec.Username
		} else if ref := basicAuthSpec.UsernameRef; ref != nil {
			if handler == nil {
				b.err = errors.Errorf("references handler is nil")
				return
			}
			username = converter.UnsafeBytesToString(handler.GetData(ref.Name, ref.Item))
		}

		var password string
		if basicAuthSpec.Password != "" {
			password = basicAuthSpec.Password
		} else if ref := basicAuthSpec.PasswordRef; ref != nil {
			if handler == nil {
				b.err = errors.Errorf("references handler is nil")
				return
			}
			password = converter.UnsafeBytesToString(handler.GetData(ref.Name, ref.Item))
		}

		if username == "" || password == "" {
			b.err = errors.Errorf("illegal basic auth account as blank username or password")
			return
		}
		status.SetUsername(username).SetPassword(password)
	}

	// processes TLS
	if clientSpec.TLSConfig != nil {
		var tlsConfigSpec = clientSpec.TLSConfig

		var caPool, caPoolErr = func() (*x509.CertPool, error) {
			var caPEM []byte
			if tlsConfigSpec.CAFilePEM != "" {
				caPEM = converter.UnsafeStringToBytes(tlsConfigSpec.CAFilePEM)
			} else if ref := tlsConfigSpec.CAFilePEMRef; ref != nil {
				if handler == nil {
					return nil, errors.Errorf("references handler is nil")
				}
				caPEM = handler.GetData(ref.Name, ref.Item)
			}
			if len(caPEM) == 0 {
				return nil, errors.Errorf("illegal TLS/SSL configuration as blank CA file")
			}

			var caPool = x509.NewCertPool()
			_ = caPool.AppendCertsFromPEM(caPEM)
			return caPool, nil
		}()
		if caPoolErr != nil {
			b.err = caPoolErr
			return
		}

		var certs, certsErr = func() ([]tls.Certificate, error) {
			var certPEM []byte
			if tlsConfigSpec.CertFilePEM != "" {
				certPEM = converter.UnsafeStringToBytes(tlsConfigSpec.CertFilePEM)
			} else if ref := tlsConfigSpec.CertFilePEMRef; ref != nil {
				if handler == nil {
					return nil, errors.Errorf("references handler is nil")
				}
				certPEM = handler.GetData(ref.Name, ref.Item)
			}
			if len(certPEM) == 0 {
				return nil, nil
			}

			var keyPEM []byte
			if tlsConfigSpec.KeyFilePEM != "" {
				keyPEM = converter.UnsafeStringToBytes(tlsConfigSpec.KeyFilePEM)
			} else if ref := tlsConfigSpec.KeyFilePEMRef; ref != nil {
				if handler == nil {
					return nil, errors.Errorf("references handler is nil")
				}
				keyPEM = handler.GetData(ref.Name, ref.Item)
			}
			if len(keyPEM) == 0 {
				return nil, nil
			}

			var certs []tls.Certificate
			var cert, err = tls.X509KeyPair(certPEM, keyPEM)
			if err != nil {
				return nil, errors.Wrap(err, "failed to construct client X509 key pair")
			}
			certs = append(certs, cert)
			return certs, nil
		}()
		if certsErr != nil {
			b.err = certsErr
			return
		}

		status.SetTLSConfig(&tls.Config{
			RootCAs:            caPool,
			Certificates:       certs,
			ServerName:         tlsConfigSpec.ServerName,
			InsecureSkipVerify: tlsConfigSpec.InsecureSkipVerify,
		})
	}

	// processes storage
	if clientSpec.Store != nil {
		status.SetStore(func() mqtt.Store {
			if clientSpec.Store.Type == "File" {
				var directoryPrefix = "/var/run/octopus/mqtt"
				if clientSpec.Store.DirectoryPrefix != "" {
					directoryPrefix = clientSpec.Store.DirectoryPrefix
				}
				var directory = filepath.Join(filepath.FromSlash(directoryPrefix), string(ref.UID))
				return mqtt.NewFileStore(directory)
			}
			return mqtt.NewMemoryStore()
		}())
	}

	// processes client id
	status.SetClientID(func() string {
		var clientIDPrefix = "octopus-"
		var clientIDSuffixLength = 32
		if clientSpec.ProtocolVersion != nil && *clientSpec.ProtocolVersion == 3 {
			// uses MQTT v3.1
			clientIDSuffixLength = 23 - len(clientIDPrefix)
		}
		var clientID = fmt.Sprintf("%s%s", clientIDPrefix, uuid.Truncate(string(ref.UID), clientIDSuffixLength))
		return clientID
	}())

	// processes will message
	if messageSpec.Will != nil {
		var topic = messageSpec.Will.Topic
		if topic == "" {
			topic = path.Join(messageSpec.Topic, "$will")
		}
		var willTopic = NewSegmentTopic(topic, messageSpec.MQTTMessageTopicOperation, ref).RenderForPublish()
		status.SetBinaryWill(willTopic, messageSpec.Will.Content.Data, 1, true)
	}

	// processes other properties
	if clientSpec.Server != "" {
		status.AddBroker(clientSpec.Server)
	}
	if clientSpec.ProtocolVersion != nil {
		status.SetProtocolVersion(*clientSpec.ProtocolVersion)
	}
	if clientSpec.CleanSession != nil {
		status.SetCleanSession(*clientSpec.CleanSession)
	}
	if clientSpec.ResumeSubs != nil {
		status.SetResumeSubs(*clientSpec.ResumeSubs)
	}
	if clientSpec.ConnectTimeout != nil {
		status.SetConnectTimeout(clientSpec.ConnectTimeout.Duration)
	}
	if clientSpec.KeepAlive != nil {
		status.SetKeepAlive(clientSpec.KeepAlive.Duration)
	}
	if clientSpec.PingTimeout != nil {
		status.SetPingTimeout(clientSpec.PingTimeout.Duration)
	}
	if clientSpec.Order != nil {
		status.SetOrderMatters(*clientSpec.Order)
	}
	if clientSpec.WriteTimeout != nil {
		status.SetWriteTimeout(clientSpec.WriteTimeout.Duration)
	}
	if clientSpec.AutoReconnect != nil {
		status.SetAutoReconnect(*clientSpec.AutoReconnect)
	}
	if clientSpec.MaxReconnectInterval != nil {
		status.SetMaxReconnectInterval(clientSpec.MaxReconnectInterval.Duration)
	}
	if clientSpec.MessageChannelDepth != nil {
		status.SetMessageChannelDepth(*clientSpec.MessageChannelDepth)
	}
	if len(clientSpec.HTTPHeaders) != 0 {
		status.SetHTTPHeaders(clientSpec.HTTPHeaders)
	}
}

// GetOptions returns the MQTT client options, call it after called `Render`.
func (b *ClientBuilder) GetOptions() *mqtt.ClientOptions {
	return b.status
}

// ConfigureOptions allows to customize the MQTT client options.
func (b *ClientBuilder) ConfigureOptions(customFunc CustomMQTTOptionFunc) {
	if b.err == nil && customFunc != nil {
		b.err = customFunc(b.status)
	}
}

// Build returns a MQTT client wrapper.
func (b *ClientBuilder) Build() (Client, error) {
	if b.err != nil {
		return nil, b.err
	}

	var ref = b.ref
	var clientSpec = b.spec.Client
	var messageSpec = b.spec.Message
	var status = b.status

	// constructs client
	var cli = &client{
		raw:                   mqtt.NewClient(status),
		topic:                 NewSegmentTopic(messageSpec.Topic, messageSpec.MQTTMessageTopicOperation, ref),
		qos:                   1,
		retained:              true,
		subscribeTopicIndexer: SubscribeTopicIndex{},
	}
	if clientSpec.WaitTimeout != nil {
		cli.waitDuration = clientSpec.WaitTimeout.Duration
	}
	if clientSpec.DisconnectQuiesce != nil {
		cli.disconnectQuiesce = clientSpec.DisconnectQuiesce.Duration
	}
	if messageSpec.QoS != nil {
		cli.qos = byte(*messageSpec.QoS)
	}
	if messageSpec.Retained != nil {
		cli.retained = *messageSpec.Retained
	}

	log.Println("Build  ",
		fmt.Sprintf("client (%s/%s, %s)=> "+
			"server: %v, client id: %q, "+
			"tls: %v, basic auth: %v, "+
			"will topic: %q, autoconnect: %v",
			ref.Namespace, ref.Name, ref.UID,
			status.Servers, status.ClientID,
			status.TLSConfig != nil, status.Username != "" && status.Password != "",
			status.WillTopic, status.AutoReconnect,
		),
	)

	return cli, nil
}

// NewClientBuilder creates the MQTT client builder.
func NewClientBuilder(spec api.MQTTOptions, ref corev1.ObjectReference) *ClientBuilder {
	return &ClientBuilder{ref: ref, spec: &spec, status: mqtt.NewClientOptions()}
}

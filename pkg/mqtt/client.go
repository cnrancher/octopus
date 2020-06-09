package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"path/filepath"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/pkg/mqtt/api/v1alpha1"
	"github.com/rancher/octopus/pkg/util/converter"
	"github.com/rancher/octopus/pkg/util/uuid"
)

type Client interface {
	// IsConnected returns a bool signifying whether
	// the client is connected or not.
	IsConnected() bool

	// IsConnectionOpen return a bool signifying whether the client has an active
	// connection to MQTT broker, i.e not in disconnected or reconnect mode.
	IsConnectionOpen() bool

	// Connect will create a connection to the message broker, by default
	// it will attempt to connect at v3.1.1 and auto retry at v3.1 if that
	// fails.
	Connect() error

	// Disconnect will end the connection with the server, but not before waiting
	// the specified number of milliseconds to wait for existing work to be
	// completed.
	Disconnect(duration time.Duration)

	// RawClient returns the original MQTT client.
	RawClient() mqtt.Client

	// Publish will publish a message with the specified QoS and content
	// to the specified topic.
	Publish(payload interface{}) error
}

type client struct {
	client        mqtt.Client
	topic         string
	qos           byte
	retained      bool
	waitTime      time.Duration
	payloadEncode string
}

func (c *client) IsConnected() bool {
	return c.client.IsConnected()
}

func (c *client) IsConnectionOpen() bool {
	return c.client.IsConnectionOpen()
}

func (c *client) Connect() error {
	var token = c.client.Connect()
	_ = token.Wait()
	return token.Error()
}

func (c *client) Disconnect(duration time.Duration) {
	c.client.Disconnect(uint(duration.Milliseconds()))
}

func (c *client) RawClient() mqtt.Client {
	return c.client
}

func (c *client) Publish(rawPayload interface{}) error {
	if rawPayload == nil {
		return nil
	}

	var payload []byte
	switch p := rawPayload.(type) {
	case []byte:
		payload = p
	case string:
		payload = converter.UnsafeStringToBytes(p)
	default:
		var err error
		payload, err = converter.MarshalJSON(rawPayload)
		if err != nil {
			return errors.Wrap(err, "failed to marshal the payload as JSON")
		}
	}

	switch c.payloadEncode {
	case "base64":
		payload = converter.EncodeBase64(payload)
	default:
		// nothing to do
	}

	var token = c.client.Publish(c.topic, c.qos, c.retained, payload)
	if c.waitTime == 0 {
		_ = token.Wait()
	} else {
		if !token.WaitTimeout(c.waitTime) {
			return errors.Errorf("timeout in %v", c.waitTime)
		}
	}
	return token.Error()
}

func NewClient(object metav1.Object, options v1alpha1.MQTTOptionsSpec, references map[string]map[string][]byte) (Client, *v1alpha1.MQTTOptionsStatus, error) {
	if object == nil {
		return nil, nil, errors.New("invalidate runtime object")
	}

	messageOptionsOutline, err := getMessageOptionsOutline(object, options.Message)
	if err != nil {
		return nil, nil, err
	}

	clientOptionsOutline, clientOptions, err := getClientOptionsOutline(string(object.GetUID()), messageOptionsOutline.TopicName, options.Client, references)
	if err != nil {
		return nil, nil, err
	}

	var cli = &client{
		client:        mqtt.NewClient(clientOptions),
		topic:         messageOptionsOutline.TopicName,
		qos:           messageOptionsOutline.QoS,
		retained:      messageOptionsOutline.Retained,
		waitTime:      messageOptionsOutline.WaitTimeout.Duration,
		payloadEncode: messageOptionsOutline.PayloadEncode,
	}
	var outline = &v1alpha1.MQTTOptionsStatus{
		Client:  *clientOptionsOutline,
		Message: *messageOptionsOutline,
	}
	return cli, outline, nil
}

func getTopic(obj metav1.Object, topicConfig v1alpha1.MQTTMessageTopic) (string, error) {
	if topicConfig.Name != "" {
		return topicConfig.Name, nil
	}

	var prefix = topicConfig.Prefix
	if prefix == "" {
		return "", errors.New("topic prefix is required in dynamic mode")
	}

	switch topicConfig.With {
	case "", "nn":
		var namespace = obj.GetNamespace()
		var name = obj.GetName()
		if namespace == "" || name == "" {
			return "", errors.New("illegal object as blank namespace or name")
		}
		return fmt.Sprintf("%s/%s/%s", prefix, namespace, name), nil
	case "uid":
		return fmt.Sprintf("%s/%s", prefix, obj.GetUID()), nil
	default:
		return "", errors.Errorf("invalidate dynamic mode: %s", topicConfig.With)
	}
}

func getMessageOptionsOutline(object metav1.Object, messageOptions v1alpha1.MQTTMessageOptions) (*v1alpha1.MQTTMessageOptionsStatus, error) {
	var outline v1alpha1.MQTTMessageOptionsStatus

	var topic, err = getTopic(object, messageOptions.Topic)
	if err != nil {
		return nil, err
	}
	outline.TopicName = topic
	outline.PayloadEncode = string(messageOptions.PayloadEncode)
	outline.QoS = byte(messageOptions.QoS)
	if messageOptions.Retained != nil {
		outline.Retained = *messageOptions.Retained
	}
	if messageOptions.WaitTimeout != nil {
		outline.WaitTimeout = *messageOptions.WaitTimeout
	}

	return &outline, nil
}

func getClientOptionsOutline(uid, topic string, clientOptions v1alpha1.MQTTClientOptions, references map[string]map[string][]byte) (*v1alpha1.MQTTClientOptionsStatus, *mqtt.ClientOptions, error) {
	var outline v1alpha1.MQTTClientOptionsStatus
	var mqttOptions = mqtt.NewClientOptions()

	// configures broker server
	if clientOptions.Server != "" {
		mqttOptions.AddBroker(clientOptions.Server)
		outline.Server = clientOptions.Server
	}
	// configures protocol version
	if clientOptions.ProtocolVersion != nil {
		mqttOptions.SetProtocolVersion(*clientOptions.ProtocolVersion)
		outline.ProtocolVersion = clientOptions.ProtocolVersion
	}
	// configures client ID
	if clientOptions.ClientID != "" {
		if clientOptions.ProtocolVersion != nil && *clientOptions.ProtocolVersion == 3 {
			// uses MQTT v3.1
			if len(clientOptions.ClientID) > 24 {
				return nil, nil, errors.New("client ID must be no longer than 23 characters")
			}
		}
		mqttOptions.SetClientID(clientOptions.ClientID)
		outline.ClientID = clientOptions.ClientID
	} else {
		var clientIDPrefix = "octopus-"
		var clientIDSuffixLength = 32
		if clientOptions.ProtocolVersion != nil && *clientOptions.ProtocolVersion == 3 {
			// uses MQTT v3.1
			clientIDSuffixLength = 23 - len(clientIDPrefix)
		}
		var clientID = fmt.Sprintf("%s%s", clientIDPrefix, uuid.Truncate(uid, clientIDSuffixLength))
		mqttOptions.SetClientID(clientID)
		outline.ClientID = clientID
	}
	// configures will
	if clientOptions.Will != nil {
		var willOutline, err = configureWill(mqttOptions, *clientOptions.Will, topic)
		if err != nil {
			return nil, nil, err
		}
		outline.Will = willOutline
	}
	// configures basic auth for connection
	if clientOptions.BasicAuth != nil {
		if err := configureBasicAuth(mqttOptions, *clientOptions.BasicAuth, references); err != nil {
			return nil, nil, err
		}
		outline.ConfigBasicAuth = true
	}
	// configures SSL/TLS for connection
	if clientOptions.TLSConfig != nil {
		if err := configureTLS(mqttOptions, *clientOptions.TLSConfig, references); err != nil {
			return nil, nil, err
		}
		outline.ConfigTLS = true
	}
	// configures clean session
	if clientOptions.CleanSession != nil {
		mqttOptions.SetCleanSession(*clientOptions.CleanSession)
		outline.CleanSession = clientOptions.CleanSession
	}
	// configures storage for persisting messages
	if clientOptions.Store != nil {
		var storeOutline, err = configureStore(mqttOptions, *clientOptions.Store, uid)
		if err != nil {
			return nil, nil, err
		}
		outline.Store = *storeOutline
	}
	// configures resume subscribe message
	if clientOptions.ResumeSubs != nil {
		mqttOptions.SetResumeSubs(*clientOptions.ResumeSubs)
		outline.ResumeSubs = clientOptions.ResumeSubs
	}
	// configures connect timeout
	if clientOptions.ConnectTimeout != nil {
		mqttOptions.SetConnectTimeout(clientOptions.ConnectTimeout.Duration)
		outline.ConnectTimeout = clientOptions.ConnectTimeout
	}
	// configures keep alive
	if clientOptions.KeepAlive != nil {
		mqttOptions.SetKeepAlive(clientOptions.KeepAlive.Duration)
		outline.KeepAlive = clientOptions.KeepAlive
	}
	// configures ping timeout
	if clientOptions.PingTimeout != nil {
		mqttOptions.SetPingTimeout(clientOptions.PingTimeout.Duration)
		outline.PingTimeout = clientOptions.PingTimeout
	}
	// configures order
	if clientOptions.Order != nil {
		mqttOptions.SetOrderMatters(*clientOptions.Order)
		outline.Order = clientOptions.Order
	}
	// configures write timeout
	if clientOptions.WriteTimeout != nil {
		mqttOptions.SetWriteTimeout(clientOptions.WriteTimeout.Duration)
		outline.WriteTimeout = clientOptions.WriteTimeout
	}
	// configures auto reconnect
	if clientOptions.AutoReconnect != nil {
		mqttOptions.SetAutoReconnect(*clientOptions.AutoReconnect)
		outline.AutoReconnect = clientOptions.AutoReconnect
	}
	// configures max reconnect interval
	if clientOptions.MaxReconnectInterval != nil {
		mqttOptions.SetMaxReconnectInterval(clientOptions.MaxReconnectInterval.Duration)
		outline.MaxReconnectInterval = clientOptions.MaxReconnectInterval
	}
	// configures message channel depth
	if clientOptions.MessageChannelDepth != nil {
		mqttOptions.SetMessageChannelDepth(*clientOptions.MessageChannelDepth)
		outline.MessageChannelDepth = clientOptions.MessageChannelDepth
	}
	// configures http headers
	if len(clientOptions.HTTPHeaders) != 0 {
		mqttOptions.SetHTTPHeaders(clientOptions.HTTPHeaders)
		outline.HTTPHeaders = clientOptions.HTTPHeaders
	}

	return &outline, mqttOptions, nil
}

func configureWill(options *mqtt.ClientOptions, willConfig v1alpha1.MQTTClientWillMessage, topic string) (*v1alpha1.MQTTClientWillMessageStatus, error) {
	var outline v1alpha1.MQTTClientWillMessageStatus

	outline.TopicName = func() string {
		if willConfig.Topic != nil {
			return willConfig.Topic.Name
		}
		return topic + "/$will"
	}()

	var payload []byte
	switch willConfig.PayloadEncode {
	case "base64":
		outline.PayloadEncode = "base64"

		var err error
		payload, err = converter.DecodeBase64String(willConfig.PayloadContent)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode base64 payload content")
		}
	default:
		payload = converter.UnsafeStringToBytes(willConfig.PayloadContent)
	}

	outline.QoS = byte(willConfig.QoS)
	if willConfig.Retained != nil {
		outline.Retained = *willConfig.Retained
	}

	options.SetBinaryWill(outline.TopicName, payload, outline.QoS, outline.Retained)
	return &outline, nil
}

func configureBasicAuth(options *mqtt.ClientOptions, basicAuthConfig v1alpha1.MQTTClientBasicAuth, references map[string]map[string][]byte) error {
	var username string
	if basicAuthConfig.Username != "" {
		username = basicAuthConfig.Username
	} else if ref := basicAuthConfig.UsernameRef; ref != nil {
		var data = referencesHandler(references).GetData(ref)
		username = converter.UnsafeBytesToString(data)
	}

	var password string
	if basicAuthConfig.Password != "" {
		password = basicAuthConfig.Password
	} else if ref := basicAuthConfig.PasswordRef; ref != nil {
		var data = referencesHandler(references).GetData(ref)
		password = converter.UnsafeBytesToString(data)
	}

	if username == "" || password == "" {
		return errors.Errorf("illegal basic auth account as blank username or password")
	}

	options.SetUsername(username).SetPassword(password)
	return nil
}

func configureTLS(options *mqtt.ClientOptions, tlsConfig v1alpha1.MQTTClientTLS, references map[string]map[string][]byte) error {
	var caPool, caPoolErr = func() (*x509.CertPool, error) {
		var caPEM []byte
		if tlsConfig.CAFilePEM != "" {
			caPEM = converter.UnsafeStringToBytes(tlsConfig.CAFilePEM)
		} else if ref := tlsConfig.CAFilePEMRef; ref != nil {
			caPEM = referencesHandler(references).GetData(ref)
		}
		if len(caPEM) == 0 {
			return nil, errors.Errorf("illegal TLS/SSL configuration as blank CA file")
		}

		var caPool = x509.NewCertPool()
		_ = caPool.AppendCertsFromPEM(caPEM)
		return caPool, nil
	}()
	if caPoolErr != nil {
		return caPoolErr
	}

	var certs, certsErr = func() ([]tls.Certificate, error) {
		var certPEM []byte
		if tlsConfig.CertFilePEM != "" {
			certPEM = converter.UnsafeStringToBytes(tlsConfig.CertFilePEM)
		} else if ref := tlsConfig.CertFilePEMRef; ref != nil {
			certPEM = referencesHandler(references).GetData(ref)
		}
		if len(certPEM) == 0 {
			return nil, nil
		}

		var keyPEM []byte
		if tlsConfig.KeyFilePEM != "" {
			keyPEM = converter.UnsafeStringToBytes(tlsConfig.KeyFilePEM)
		} else if ref := tlsConfig.KeyFilePEMRef; ref != nil {
			keyPEM = referencesHandler(references).GetData(ref)
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
		return certsErr
	}

	var config = &tls.Config{
		RootCAs:            caPool,
		Certificates:       certs,
		ServerName:         tlsConfig.ServerName,
		InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
	}
	options.SetTLSConfig(config)
	return nil
}

func configureStore(options *mqtt.ClientOptions, storeConfig v1alpha1.MQTTClientStore, uid string) (*v1alpha1.MQTTClientStoreStatus, error) {
	var outline v1alpha1.MQTTClientStoreStatus
	var store mqtt.Store

	switch storeConfig.Type {
	case "file":
		var directoryPrefix = "/var/run/octopus/mqtt"
		if storeConfig.DirectoryPrefix != "" {
			directoryPrefix = storeConfig.DirectoryPrefix
		}

		var directory = filepath.Join(filepath.FromSlash(directoryPrefix), uid)
		store = mqtt.NewFileStore(directory)

		outline.Type = "file"
		outline.Directory = directory
	default:
		store = mqtt.NewMemoryStore()
	}

	options.SetStore(store)
	return &outline, nil
}

type referencesHandler map[string]map[string][]byte

// GetData returns the data of specified name and itemName,
// it's always return nil if the data bytes is not existed or empty.
func (l1m referencesHandler) GetData(relationship *edgev1alpha1.DeviceLinkReferenceRelationship) []byte {
	if len(l1m) == 0 {
		return nil
	}

	// parses level 1
	var l2m, l2Exist = l1m[relationship.Name]
	if !l2Exist || len(l2m) == 0 {
		return nil
	}

	// parses level 2
	return l2m[relationship.Item]
}

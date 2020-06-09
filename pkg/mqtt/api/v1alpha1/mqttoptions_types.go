package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

// MQTTMessageQoSLevel defines the QoS level of publishing message.
// 0: Send at most once.
// 1: Send at least once.
// 2: Send exactly once.
// +kubebuilder:validation:Enum=0;1;2
type MQTTMessageQoSLevel byte

// MQTTMessagePayloadEncode defines the encoded method of publishing message.
// raw: Not encode.
// base64: The output (published) data is encoded with Base64, and the input (subscribed) data is decoded with Base64.
// +kubebuilder:validation:Enum=raw;base64
type MQTTMessagePayloadEncode string

// MQTTClientBasicAuth defines the basic authentication information.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTClientBasicAuth struct {
	// Specifies the username for basic authentication.
	// +optional
	Username string `json:"username,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the username.
	// +optional
	UsernameRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"usernameRef,omitempty"`

	// Specifies the password for basic authenication.
	// +optional
	Password string `json:"password,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the password.
	// +optional
	PasswordRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"passwordRef,omitempty"`
}

// MQTTClientTLS defines the SSL/TLS connection information.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTClientTLS struct {
	// Specifies the PEM format content of the CA certificate,
	// which is used for validate the server certificate with.
	// +optional
	CAFilePEM string `json:"caFilePEM,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the CA file PEM content.
	// +optional
	CAFilePEMRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"caFilePEMRef,omitempty"`

	// Specifies the PEM format content of the certificate(public key),
	// which is used for client authenticate to the server.
	// +optional
	CertFilePEM string `json:"certFilePEM,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the client certificate file PEM content.
	// +optional
	CertFilePEMRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"certFilePEMRef,omitempty"`

	// Specifies the PEM format content of the key(private key),
	// which is used for client authenticate to the server.
	// +optional
	KeyFilePEM string `json:"keyFilePEM,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the client key file PEM content.
	// +optional
	KeyFilePEMRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"keyFilePEMRef,omitempty"`

	// Indicates the name of the server,
	// ref to http://tools.ietf.org/html/rfc4366#section-3.1.
	// +optional
	ServerName string `json:"serverName,omitempty"`

	// Doesn't validate the server certificate.
	// +optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// MQTTClientStorageType defines the storage type for persisting messages.
// memory: Stores in memory.
// file: Stores in files.
// +kubebuilder:validation:Enum=memory;file
type MQTTClientStorageType string

// MQTTClientStore is the Schema for configuring the MQTT client store.
type MQTTClientStore struct {
	// Specifies the type of storage.
	// The default value is "memory".
	// +optional
	Type MQTTClientStorageType `json:"type,omitempty"`

	// Specifies the directory prefix of the storage, if using file store.
	// The default value is "/var/run/octopus/mqtt".
	// +kubebuilder:validation:Pattern="^/.*[^/]$"
	// +optional
	DirectoryPrefix string `json:"directoryPrefix,omitempty"`
}

// MQTTMessageTopicStatic is the Schema for configuring the MQTT broker static topic.
type MQTTMessageTopicStatic struct {
	// Specifies the static name of topic.
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9_]+.*[^/]$"
	// +optional
	Name string `json:"name,omitempty"`
}

// MQTTMessageTopicDynamicWith defines the suffix source of the dynamic name.
// nn: Append with the NamespacedName of the runtime object.
// uid: Append with the UID of the runtime object.
// +kubebuilder:validation:Enum=uid;nn
type MQTTMessageTopicDynamicWith string

// MQTTMessageTopicDynamic is the Schema for configuring the MQTT broker dynamic topic.
type MQTTMessageTopicDynamic struct {
	// Specifies the prefix for the dynamic name of topic.
	// The prefix is required for dynamic name.
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9_]+.*[^/]$"
	// +optional
	Prefix string `json:"prefix,omitempty"`

	// Specifies the mode for the dynamic name of topic.
	// The default mode is "nn".
	// +optional
	With MQTTMessageTopicDynamicWith `json:"with,omitempty"`
}

// MQTTMessageTopic is the Schema for configuring the MQTT broker topic.
type MQTTMessageTopic struct {
	MQTTMessageTopicStatic  `json:",inline"`
	MQTTMessageTopicDynamic `json:",inline"`
}

// MQTTClientWillMessage is the Schema for configuring the will of MQTT client.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTClientWillMessage struct {
	// Specifies the topic for publishing the will message,
	// if not set, the will topic will append "$will" to the topic name specified
	// in global settings as its topic name.
	// +optional
	Topic *MQTTMessageTopicStatic `json:"topic,omitempty"`

	// Specifies the encode way of payload content.
	// The "base64" way allows to input bytes (`PayloadContent`) that cannot be characterized.
	// The default way is "raw".
	// +optional
	PayloadEncode MQTTMessagePayloadEncode `json:"payloadEncode,omitempty"`

	// Specifies the payload content.
	// +kubebuilder:validation:Required
	PayloadContent string `json:"payloadContent"`

	// Specifies the QoS of the will message.
	// The default value is "0".
	// +optional
	QoS MQTTMessageQoSLevel `json:"qos,omitempty"`

	// Specifies the will message to be retained.
	// The default value is "false".
	// +optional
	Retained *bool `json:"retained,omitempty"`
}

// MQTTClientOptions is the Schema for configuring the MQTT client.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTClientOptions struct {
	// Specifies the server URI of MQTT broker, the format should be `schema://host:port`.
	// The "schema" is one of the "ws", "wss", "tcp", "unix", "ssl", "tls" or "tcps".
	// +kubebuilder:validation:Pattern="^(ws|wss|tcp|unix|ssl|tls|tcps)+://[^\\s]*$"
	// +kubebuilder:validation:Required
	Server string `json:"server"`

	// Specifies the MQTT protocol version that the cluster uses to connect to broker.
	// Legitimate values are currently 3 - MQTT v3.1 or 4 - MQTT v3.1.1.
	// +kubebuilder:validation:Enum=3;4
	// +optional
	ProtocolVersion *uint `json:"protocolVersion,omitempty"`

	// Specifies the client ID to be used for connection, it must be no longer than 23 characters
	// if specified to use MQTT v3.1.
	// +optional
	ClientID string `json:"clientID"`

	// Specifies the will message that the client gives it to the broker,
	// which can be published to any clients that are subscribed the provided topic.
	// +optional
	Will *MQTTClientWillMessage `json:"will,omitempty"`

	// Specifies the username and password that the client connects
	// to the MQTT broker. Without the use of TLSConfig,
	// the account information will be sent in plaintext across the wire.
	// +optional
	BasicAuth *MQTTClientBasicAuth `json:"basicAuth,omitempty"`

	// Specifies the TLS configuration that the client connects to the MQTT broker.
	// +optional
	TLSConfig *MQTTClientTLS `json:"tlsConfig,omitempty"`

	// Specifies setting the "clean session" flag in the connect message that
	// the MQTT broker should not save it.
	// If the value is "false", the broker stores all missed messages
	// for the client that subscribed with QoS 1 or 2.
	// Any messages that were going to be sent by this client
	// before disconnecting previously but didn't send upon connecting to the broker.
	// The default value is "true".
	CleanSession *bool `json:"cleanSession,omitempty"`

	// Specifies to provide message persistence in cases where QoS level is 1 or 2.
	// The default store is "memory".
	// +optional
	Store *MQTTClientStore `json:"store,omitempty"`

	// Specifies to enable resuming of stored (un)subscribe messages when connecting but not reconnecting.
	// This is only valid if `CleanSession` is false.
	// The default value is "false".
	// +optional
	ResumeSubs *bool `json:"resumeSubs,omitempty"`

	// Specifies the amount of time that the client try to open a connection
	// to an MQTT broker before timing out and getting error.
	// A duration of 0 never times out.
	// The default value is "30s".
	// +optional
	ConnectTimeout *metav1.Duration `json:"connectTimeout,omitempty"`

	// Specifies the amount of time that the client should wait
	// before sending a PING request to the broker. This will
	// allow the client to know that the connection has not been lost
	// with the server.
	// A duration of 0 never keeps alive.
	// The default keep alive is "30s".
	// +optional
	KeepAlive *metav1.Duration `json:"keepAlive,omitempty"`

	// Specifies the amount of time that the client should wait
	// after sending a PING request to the broker. This will
	// allow the client to know that the connection has been lost
	// with the server.
	// A duration of 0 may cause unnecessary timeout error.
	// The default value is "10s".
	// +optional
	PingTimeout *metav1.Duration `json:"pingTimeout,omitempty"`

	// Specifies the message routing to guarantee order within each QoS level. If set to false,
	// the message can be delivered asynchronously from the client to the application and
	// possibly arrive out of order.
	// The default value is "true".
	// +optional
	Order *bool `json:"order,omitempty"`

	// Specifies the amount of time that the client publish a message successfully before
	// getting a timeout error.
	// A duration of 0 never times out.
	// The default value is "30s".
	// +optional
	WriteTimeout *metav1.Duration `json:"writeTimeout,omitempty"`

	// Configures using the automatic reconnection logic.
	// The default value is "true".
	// +optional
	AutoReconnect *bool `json:"autoReconnect,omitempty"`

	// Specifies the amount of time that the client should wait
	// before reconnecting to the broker. The first reconnect interval is 1 second,
	// and then the interval is incremented by *2 until `MaxReconnectInterval` is reached.
	// This is only valid if `AutoReconnect` is true.
	// A duration of 0 may trigger the reconnection immediately.
	// The default value is "10m".
	// +optional
	MaxReconnectInterval *metav1.Duration `json:"maxReconnectInterval,omitempty"`

	// Specifies the size of the internal queue that holds messages
	// while the client is temporarily offline, allowing the application to publish
	// when the client is reconnected.
	// This is only valid if `AutoReconnect` is true.
	// The default value is "100".
	// +optional
	MessageChannelDepth *uint `json:"messageChannelDepth,omitempty"`

	// Specifies the additional HTTP headers that the client sends in the WebSocket opening handshake.
	// +optional
	HTTPHeaders map[string][]string `json:"httpHeaders,omitempty"`
}

// MQTTMessagePayloadOptions is the Schema for configuring the MQTT message payload.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTMessagePayloadOptions struct {
	// Specifies the encode way of payload content.
	// The default way is "raw".
	// +optional
	PayloadEncode MQTTMessagePayloadEncode `json:"payloadEncode,omitempty"`

	// Specifies the QoS of the message.
	// The default value is "0".
	// +optional
	QoS MQTTMessageQoSLevel `json:"qos,omitempty"`

	// Specifies if the last published message to be retained.
	// The default value is "false".
	// +optional
	Retained *bool `json:"retained,omitempty"`

	// Specifies the amount of time that the client should wait
	// after operating.
	// A duration of 0 never times out.
	// The default value is "0s".
	// +optional
	WaitTimeout *metav1.Duration `json:"waitTimeout,omitempty"`
}

// MQTTMessageOptions is the Schema for configuring the MQTT message.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTMessageOptions struct {
	MQTTMessagePayloadOptions `json:",inline"`

	// Specifies the topic settings.
	// +kubebuilder:validation:Required
	Topic MQTTMessageTopic `json:"topic"`
}

// MQTTOptionsSpec defines the desired value of MQTT client.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTOptionsSpec struct {
	// Specifies the client settings.
	// +kubebuilder:validation:Required
	Client MQTTClientOptions `json:"client"`

	// Specifies the message settings.
	// +kubebuilder:validation:Required
	Message MQTTMessageOptions `json:"message"`
}

// MQTTClientWillMessageStatus defines the observed state of MQTT client will message options.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTClientWillMessageStatus struct {
	// Observes the topic for publishing the will message.
	// +optional
	TopicName string `json:"topicName,omitempty"`

	// Observes the encode way of payload content.
	// +kubebuilder:default="raw"
	PayloadEncode string `json:"payloadEncode,omitempty"`

	// Observes the QoS of the will message.
	// +kubebuilder:default=0
	QoS byte `json:"qos,omitempty"`

	// Observes if retaining the will message.
	// +kubebuilder:default=false
	Retained bool `json:"retained,omitempty"`
}

// MQTTClientStoreStatus defines the observed state of MQTT store options.
type MQTTClientStoreStatus struct {
	// Observes the type of storage.
	// +kubebuilder:default=memory
	// +optional
	Type string `json:"type,omitempty"`

	// Observes the directory of the file storage.
	// +optional
	Directory string `json:"directory,omitempty"`
}

// MQTTClientOptionsStatus defines the observed state of MQTT client options.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTClientOptionsStatus struct {
	// Observes the broker server URI.
	// +optional
	Server string `json:"server,omitempty"`

	// Observes the protocol version.
	// +optional
	ProtocolVersion *uint `json:"protocolVersion,omitempty"`

	// Observes the client ID.
	// +optional
	ClientID string `json:"clientID,omitempty"`

	// Observes the will message that the client gives it to the broker.
	// +optional
	Will *MQTTClientWillMessageStatus `json:"will,omitempty"`

	// Observes if configuring basic authentication.
	// +optional
	ConfigBasicAuth bool `json:"configBasicAuth,omitempty"`

	// Observes if configuring TLS.
	// +optional
	ConfigTLS bool `json:"configTLS,omitempty"`

	// Observes if setting the "clean session" flag.
	// +kubebuilder:default=true
	// +optional
	CleanSession *bool `json:"cleanSession,omitempty"`

	// Observes the store type.
	// +optional
	Store MQTTClientStoreStatus `json:"store,omitempty"`

	// Observes if enabling resuming of stored (un)subscribe messages when connecting but not reconnecting.
	// +kubebuilder:default=false
	ResumeSubs *bool `json:"resumeSubs,omitempty"`

	// Observes the amount of time that the client try to open a connection
	// to an MQTT broker before timing out and getting error.
	// +kubebuilder:default="30s"
	ConnectTimeout *metav1.Duration `json:"connectTimeout,omitempty"`

	// Observes the amount of time that the client should wait
	// before sending a PING request to the broker.
	// +kubebuilder:default="30s"
	KeepAlive *metav1.Duration `json:"keepAlive,omitempty"`

	// Observes the amount of time that the client should wait
	// after sending a PING request to the broker.
	// +kubebuilder:default="10s"
	PingTimeout *metav1.Duration `json:"pingTimeout,omitempty"`

	// Observes the message routing to guarantee order within each QoS level.
	// +kubebuilder:default=true
	Order *bool `json:"order,omitempty"`

	// Observes the amount of time that the client publish a message successfully before
	// getting a timeout error.
	// +kubebuilder:default="0s"
	WriteTimeout *metav1.Duration `json:"writeTimeout,omitempty"`

	// Observes if using the automatic reconnection logic.
	// +kubebuilder:default=true
	AutoReconnect *bool `json:"autoReconnect,omitempty"`

	// Observes the amount of time that the client should wait
	// before reconnecting to the broker.
	// +kubebuilder:default="10m"
	MaxReconnectInterval *metav1.Duration `json:"maxReconnectInterval,omitempty"`

	// Observes the size of the internal queue that holds messages
	// while the client is temporarily offline, allowing the application to publish
	// when the client is reconnected.
	// +kubebuilder:default=100
	MessageChannelDepth *uint `json:"messageChannelDepth,omitempty"`

	// Observes the additional HTTP headers that the client sends in the WebSocket opening handshake.
	// +optional
	HTTPHeaders map[string][]string `json:"httpHeaders,omitempty"`
}

// MQTTMessageOptionsStatus defines the observed state of MQTT message options.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTMessageOptionsStatus struct {
	// Observes the topic for publishing/subscribing the message.
	// +optional
	TopicName string `json:"topicName,omitempty"`

	// Observes the encode way of payload content.
	// +kubebuilder:default="raw"
	PayloadEncode string `json:"payloadEncode,omitempty"`

	// Observes the QoS of the message.
	// +kubebuilder:default=0
	QoS byte `json:"qos,omitempty"`

	// Observes if retaining the message.
	// +kubebuilder:default=false
	Retained bool `json:"retained,omitempty"`

	// Observes the amount of time that the client should wait
	// after operating.
	// +kubebuilder:default="0s"
	WaitTimeout metav1.Duration `json:"waitTimeout,omitempty"`
}

// MQTTOptionsStatus defines the observed options of the MQTT.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTOptionsStatus struct {
	// Observes the client settings.
	// +optional
	Client MQTTClientOptionsStatus `json:"client,omitempty"`

	// Observes the message settings.
	// +optional
	Message MQTTMessageOptionsStatus `json:"message,omitempty"`
}

package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

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
// Memory: Stores in memory.
// File: Stores in files.
// +kubebuilder:validation:Enum=Memory;File
type MQTTClientStorageType string

// MQTTClientStore defines the storage of MQTT client.
type MQTTClientStore struct {
	// Specifies the type of storage.
	// The default value is "Memory".
	// +kubebuilder:default="Memory"
	// +optional
	Type MQTTClientStorageType `json:"type,omitempty"`

	// Specifies the directory prefix of the storage, if using file store.
	// The default value is "/var/run/octopus/mqtt".
	// +kubebuilder:validation:Pattern="^/.*[^/]$"
	// +optional
	DirectoryPrefix string `json:"directoryPrefix,omitempty"`
}

// MQTTClientOptions defines the options of MQTT client.
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
	// The default value is 0, which means MQTT v3.1.1 identification is preferred.
	// +kubebuilder:validation:Enum=0;3;4
	// +kubebuilder:default=0
	// +optional
	ProtocolVersion *uint `json:"protocolVersion,omitempty"`

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
	// +kubebuilder:default=true
	CleanSession *bool `json:"cleanSession,omitempty"`

	// Specifies to provide message persistence in cases where QoS level is 1 or 2.
	// +optional
	Store *MQTTClientStore `json:"store,omitempty"`

	// Specifies to enable resuming of stored (un)subscribe messages when connecting but not reconnecting.
	// This is only valid if `CleanSession` is false.
	// The default value is "false".
	// +kubebuilder:default=false
	// +optional
	ResumeSubs *bool `json:"resumeSubs,omitempty"`

	// Specifies the amount of time that the client try to open a connection
	// to an MQTT broker before timing out and getting error.
	// A duration of 0 never times out.
	// The default value is "30s".
	// +kubebuilder:default="30s"
	// +optional
	ConnectTimeout *metav1.Duration `json:"connectTimeout,omitempty"`

	// Specifies the amount of time that the client should wait
	// before sending a PING request to the broker. This will
	// allow the client to know that the connection has not been lost
	// with the server.
	// A duration of 0 never keeps alive.
	// The default keep alive is "30s".
	// +kubebuilder:default="30s"
	// +optional
	KeepAlive *metav1.Duration `json:"keepAlive,omitempty"`

	// Specifies the amount of time that the client should wait
	// after sending a PING request to the broker. This will
	// allow the client to know that the connection has been lost
	// with the server.
	// A duration of 0 may cause unnecessary timeout error.
	// The default value is "10s".
	// +kubebuilder:default="10s"
	// +optional
	PingTimeout *metav1.Duration `json:"pingTimeout,omitempty"`

	// Specifies the message routing to guarantee order within each QoS level. If set to false,
	// the message can be delivered asynchronously from the client to the application and
	// possibly arrive out of order.
	// The default value is "true".
	// +kubebuilder:default=true
	// +optional
	Order *bool `json:"order,omitempty"`

	// Specifies the amount of time that the client publish a message successfully before
	// getting a timeout error.
	// A duration of 0 never times out.
	// The default value is "30s".
	// +kubebuilder:default="30s"
	// +optional
	WriteTimeout *metav1.Duration `json:"writeTimeout,omitempty"`

	// Specifies the amount of time that the client should timeout
	// after subscribed/published a message.
	// A duration of 0 never times out.
	// +optional
	WaitTimeout *metav1.Duration `json:"waitTimeout,omitempty"`

	// Specifies the quiesce when the client disconnects.
	// The default value is "5s".
	// +optional
	DisconnectQuiesce *metav1.Duration `json:"disconnectQuiesce,omitempty"`

	// Configures using the automatic reconnection logic.
	// The default value is "true".
	// +kubebuilder:default=true
	// +optional
	AutoReconnect *bool `json:"autoReconnect,omitempty"`

	// Specifies the amount of time that the client should wait
	// before reconnecting to the broker. The first reconnect interval is 1 second,
	// and then the interval is incremented by *2 until `MaxReconnectInterval` is reached.
	// This is only valid if `AutoReconnect` is true.
	// A duration of 0 may trigger the reconnection immediately.
	// The default value is "10m".
	// +kubebuilder:default="10m"
	// +optional
	MaxReconnectInterval *metav1.Duration `json:"maxReconnectInterval,omitempty"`

	// Specifies the size of the internal queue that holds messages
	// while the client is temporarily offline, allowing the application to publish
	// when the client is reconnected.
	// This is only valid if `AutoReconnect` is true.
	// The default value is "100".
	// +kubebuilder:default=100
	// +optional
	MessageChannelDepth *uint `json:"messageChannelDepth,omitempty"`

	// Specifies the additional HTTP headers that the client sends in the WebSocket opening handshake.
	// +optional
	// +mapType=atomic
	HTTPHeaders map[string][]string `json:"httpHeaders,omitempty"`
}

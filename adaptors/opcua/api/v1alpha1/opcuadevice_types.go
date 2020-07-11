package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

// OPCUADeviceParameters defines the desired parameters of OPCUADevice.
type OPCUADeviceParameters struct {
	// Specifies the amount of interval that synchronized to limb.
	// The default value is "15s".
	// +kubebuilder:default="15s"
	SyncInterval v1.Duration `json:"syncInterval,omitempty"`

	// Specifies the amount of timeout.
	// The default value is "10s".
	// +kubebuilder:default="10s"
	Timeout v1.Duration `json:"timeout,omitempty"`
}

func (in *OPCUADeviceParameters) GetSyncInterval() time.Duration {
	if in != nil {
		if duration := in.SyncInterval.Duration; duration > 0 {
			return duration
		}
	}
	return 15 * time.Second
}

func (in *OPCUADeviceParameters) GetTimeout() time.Duration {
	if in != nil {
		if duration := in.Timeout.Duration; duration > 0 {
			return duration
		}
	}
	return 10 * time.Second
}

// OPCUADeviceProtocolSecurityPolicy defines the policy of OPCUADeviceProtocol security.
// +kubebuilder:validation:Enum=None;Basic128Rsa15;Basic256;Basic256Sha256;Aes128Sha256RsaOaep;Aes256Sha256RsaPss
type OPCUADeviceProtocolSecurityPolicy string

// OPCUADeviceProtocolSecurityMode defines the model of OPCUADeviceProtocol security.
// +kubebuilder:validation:Enum=None;Sign;SignAndEncrypt
type OPCUADeviceProtocolSecurityMode string

// OPCUADeviceProtocolBasicAuth defines the basic authentication information.
type OPCUADeviceProtocolBasicAuth struct {
	// Specifies the username for accessing OPC-UA server.
	Username string `json:"username,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the username.
	// +optional
	UsernameRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"usernameRef,omitempty"`

	// Specifies the password for accessing OPC-UA server.
	Password string `json:"password,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the password.
	// +optional
	PasswordRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"passwordRef,omitempty"`
}

// OPCUADeviceProtocolTLS defines the SSL/TLS connection information
type OPCUADeviceProtocolTLS struct {
	// Specifies the PEM format content of the certificate(public key),
	// which is used for client authenticate to the OPC-UA server.
	// +optional
	CertFilePEM string `json:"certFilePEM,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the client certificate file PEM content.
	// +optional
	CertFilePEMRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"certFilePEMRef,omitempty"`

	// Specifies the PEM format content of the key(private key),
	// which is used for client authenticate to the OPC-UA server.
	// +optional
	KeyFilePEM string `json:"keyFilePEM,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the client key file PEM content.
	// +optional
	KeyFilePEMRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"keyFilePEMRef,omitempty"`
}

// OPCUADeviceProtocol defines the desired protocol of OPCUADevice.
type OPCUADeviceProtocol struct {
	// Specifies the URL of OPC-UA server endpoint,
	// which is start with "opc.tcp://".
	// +kubebuilder:validation:Pattern="^opc\\.tcp://[^\\s]*$"
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// Specifies the security policy for accessing OPC-UA server.
	// The default value is "None".
	// +kubebuilder:default="None"
	SecurityPolicy OPCUADeviceProtocolSecurityPolicy `json:"securityPolicy,omitempty"`

	// Specifies the security mode for accessing OPC-UA server.
	// The default value is "None".
	// +kubebuilder:default="None"
	SecurityMode OPCUADeviceProtocolSecurityMode `json:"securityMode,omitempty"`

	// Specifies the username and password that the client connects to OPC-UA server.
	// +optional
	BasicAuth *OPCUADeviceProtocolBasicAuth `json:"basicAuth,omitempty"`

	// Specifies the TLS configuration that the client connects to OPC-UA server.
	// +optional
	TLSConfig *OPCUADeviceProtocolTLS `json:"tlsConfig,omitempty"`
}

// OPCUADeviceProperty defines the desired property of OPCUADevice.
type OPCUADeviceProperty struct {
	// Specifies the name of property.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Specifies the description of property.
	// +optional
	Description string `json:"description,omitempty"`

	// Specifies the type of property.
	// +kubebuilder:validation:Required
	Type OPCUADevicePropertyType `json:"type"`

	// Specifies the visitor of property.
	// +kubebuilder:validation:Required
	Visitor OPCUADevicePropertyVisitor `json:"visitor"`

	// Specifies if the property is readonly.
	// The default value is "false".
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`

	// Specifies the value of property, only available in the writable property.
	// +optional
	Value string `json:"value,omitempty"`
}

// OPCUADevicePropertyType defines the type of property.
// +kubebuilder:validation:Enum=float;double;int64;int32;int16;uint64;uint32;uint16;string;boolean;byteString;datetime
type OPCUADevicePropertyType string

const (
	OPCUADevicePropertyTypeInt16      OPCUADevicePropertyType = "int16"
	OPCUADevicePropertyTypeInt32      OPCUADevicePropertyType = "int32"
	OPCUADevicePropertyTypeInt64      OPCUADevicePropertyType = "int64"
	OPCUADevicePropertyTypeUInt16     OPCUADevicePropertyType = "uint16"
	OPCUADevicePropertyTypeUInt32     OPCUADevicePropertyType = "uint32"
	OPCUADevicePropertyTypeUInt64     OPCUADevicePropertyType = "uint64"
	OPCUADevicePropertyTypeFloat      OPCUADevicePropertyType = "float"
	OPCUADevicePropertyTypeDouble     OPCUADevicePropertyType = "double"
	OPCUADevicePropertyTypeString     OPCUADevicePropertyType = "string"
	OPCUADevicePropertyTypeByteString OPCUADevicePropertyType = "byteString"
	OPCUADevicePropertyTypeBoolean    OPCUADevicePropertyType = "boolean"
	OPCUADevicePropertyTypeDatetime   OPCUADevicePropertyType = "datetime"
)

// OPCUADevicePropertyVisitor defines the visitor of property.
type OPCUADevicePropertyVisitor struct {
	// Specifies the id of OPC-UA node, e.g. "ns=1,i=1005".
	// +kubebuilder:validation:Required
	NodeID string `json:"nodeID"`

	// Specifies the name of OPC-UA node.
	// +optional
	BrowseName string `json:"browseName,omitempty"`
}

// OPCUADeviceStatusProperty defines the observed property of OPCUADevice.
type OPCUADeviceStatusProperty struct {
	// Reports the name of property.
	// +optional
	Name string `json:"name,omitempty"`

	// Reports the type of property.
	// +optional
	Type OPCUADevicePropertyType `json:"type,omitempty"`

	// Reports the value of property.
	// +optional
	Value string `json:"value,omitempty"`

	// Reports the updated timestamp of property.
	// +optional
	UpdatedAt *metav1.Time `json:"updatedAt,omitempty"`
}

// OPCUADeviceSpec defines the desired state of OPCUADevice.
type OPCUADeviceSpec struct {
	// Specifies the extension of device.
	// +optional
	Extension *OPCUADeviceExtension `json:"extension,omitempty"`

	// Specifies the parameters of device.
	// +optional
	Parameters *OPCUADeviceParameters `json:"parameters,omitempty"`

	// Specifies the protocol for accessing the device.
	// +kubebuilder:validation:Required
	Protocol OPCUADeviceProtocol `json:"protocol"`

	// Specifies the properties of device.
	// +listType=map
	// +listMapKey=name
	// +optional
	Properties []OPCUADeviceProperty `json:"properties,omitempty"`
}

// OPCUADeviceStatus defines the observed state of OPCUADevice.
type OPCUADeviceStatus struct {
	// Reports the properties of device.
	// +optional
	Properties []OPCUADeviceStatusProperty `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=opcua
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ENDPOINT",type="string",JSONPath=`.spec.protocol.endpoint`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=`.metadata.creationTimestamp`
// OPCUADevice is the schema for the OPC-UA device API.
type OPCUADevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OPCUADeviceSpec   `json:"spec,omitempty"`
	Status OPCUADeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// OPCUADeviceList contains a list of OPC-UA devices.
type OPCUADeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []OPCUADevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OPCUADevice{}, &OPCUADeviceList{})
}

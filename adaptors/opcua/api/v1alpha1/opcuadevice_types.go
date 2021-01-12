package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

// OPCUADevicePropertyAccessMode defines the mode for accessing a device property,
// default is "ReadMany".
// +kubebuilder:validation:Enum=Notify;WriteOnce;WriteMany;ReadOnce;ReadMany
type OPCUADevicePropertyAccessMode string

const (
	OPCUADevicePropertyAccessModeNotify    OPCUADevicePropertyAccessMode = "Notify"
	OPCUADevicePropertyAccessModeWriteOnce OPCUADevicePropertyAccessMode = "WriteOnce"
	OPCUADevicePropertyAccessModeWriteMany OPCUADevicePropertyAccessMode = "WriteMany"
	OPCUADevicePropertyAccessModeReadOnce  OPCUADevicePropertyAccessMode = "ReadOnce"
	OPCUADevicePropertyAccessModeReadMany  OPCUADevicePropertyAccessMode = "ReadMany"
)

// OPCUADeviceProtocolSecurityPolicy defines the policy of OPCUADeviceProtocol security.
// +kubebuilder:validation:Enum=None;Basic128Rsa15;Basic256;Basic256Sha256;Aes128Sha256RsaOaep;Aes256Sha256RsaPss
type OPCUADeviceProtocolSecurityPolicy string

// OPCUADeviceProtocolSecurityMode defines the model of OPCUADeviceProtocol security.
// +kubebuilder:validation:Enum=None;Sign;SignAndEncrypt
type OPCUADeviceProtocolSecurityMode string

// OPCUADeviceProtocolBasicAuth defines the client basic authentication information.
type OPCUADeviceProtocolBasicAuth struct {
	// Specifies the username for accessing OPC-UA server.
	// +optional
	Username string `json:"username,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the username.
	// +optional
	UsernameRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"usernameRef,omitempty"`

	// Specifies the password for accessing OPC-UA server.
	// +optional
	Password string `json:"password,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the password.
	// +optional
	PasswordRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"passwordRef,omitempty"`
}

// OPCUADeviceProtocolCertificateAuth defines the client certificate authentication information.
type OPCUADeviceProtocolCertificateAuth struct {
	// Specifies the PEM format content of the certificate(public key),
	// which is used for the client accessing OPC-UA server.
	// +optional
	CertFilePEM string `json:"certFilePEM,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the client certificate file PEM content.
	// +optional
	CertFilePEMRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"certFilePEMRef,omitempty"`
}

// OPCUADeviceProtocolIssuedTokenAuth defines the client issued token authentication information.
type OPCUADeviceProtocolIssuedTokenAuth struct {
	// Specifies the token for accessing OPC-UA server.
	// +optional
	Token string `json:"token,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the token.
	// +optional
	TokenRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"tokenRef,omitempty"`
}

// OPCUADeviceProtocolTLS defines the SSL/TLS connection information
type OPCUADeviceProtocolTLS struct {
	// Specifies the PEM format content of the certificate(public key),
	// which is used for client connect to the OPC-UA server.
	// +optional
	CertFilePEM string `json:"certFilePEM,omitempty"`

	// Specifies the relationship of DeviceLink's references to
	// refer to the value as the client certificate file PEM content.
	// +optional
	CertFilePEMRef *edgev1alpha1.DeviceLinkReferenceRelationship `json:"certFilePEMRef,omitempty"`

	// Specifies the PEM format content of the key(private key),
	// which is used for client connect to the OPC-UA server.
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

	// Specifies the username and password for the client authenticate to the OPC-UA server.
	// +optional
	BasicAuth *OPCUADeviceProtocolBasicAuth `json:"basicAuth,omitempty"`

	// Specifies the certificate for the client authenticates to the OPC-UA server.
	// +optional
	CertificateAuth *OPCUADeviceProtocolCertificateAuth `json:"certificateAuth,omitempty"`

	// Specifies the issued token for the client authenticates to the OPC-UA server.
	// +optional
	IssuedTokenAuth *OPCUADeviceProtocolIssuedTokenAuth `json:"issuedTokenAuth,omitempty"`

	// Specifies the TLS configuration for the client connect to the OPC-UA server.
	// +optional
	TLSConfig *OPCUADeviceProtocolTLS `json:"tlsConfig,omitempty"`

	// Specifies the amount of interval for synchronizing the OPC-UA server.
	// The default value is "10s".
	// +kubebuilder:default="10s"
	SyncInterval metav1.Duration `json:"syncInterval,omitempty"`

	// Specifies the amount of timeout for connecting to the OPC-UA server.
	// The default value is "10s".
	// +kubebuilder:default="10s"
	ConnectTimeout metav1.Duration `json:"connectTimeout,omitempty"`

	// Specifies the amount of timeout for sending a request to the OPC-UA server.
	// The default value is "10s".
	// +kubebuilder:default="10s"
	RequestTimeout metav1.Duration `json:"requestTimeout,omitempty"`

	// Specifies the amount of interval for renewing the connection.
	// The default value is "1h".
	// +kubebuilder:default="1h"
	RenewInterval metav1.Duration `json:"renewInterval,omitempty"`

	// Specifies the amount of interval for the OPC-UA server publishing the changes.
	// The default value is "5" seconds.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=60
	PublishingIntervalInSeconds uint8 `json:"publishingIntervalInSeconds,omitempty"`
}

func (in *OPCUADeviceProtocol) GetSyncInterval() time.Duration {
	if in != nil {
		if duration := in.SyncInterval.Duration; duration > 0 {
			return duration
		}
	}
	return 10 * time.Second
}

func (in *OPCUADeviceProtocol) GetConnectTimeout() time.Duration {
	if in != nil {
		if duration := in.ConnectTimeout.Duration; duration > 0 {
			return duration
		}
	}
	return 10 * time.Second
}

func (in *OPCUADeviceProtocol) GetRequestTimeout() time.Duration {
	if in != nil {
		if duration := in.RequestTimeout.Duration; duration > 0 {
			return duration
		}
	}
	return 10 * time.Second
}

func (in *OPCUADeviceProtocol) GetLifetime() time.Duration {
	if in != nil {
		if duration := in.RenewInterval.Duration; duration > 0 {
			return duration
		}
	}
	return 1 * time.Hour
}

func (in *OPCUADeviceProtocol) GetPublishingInterval() time.Duration {
	if in != nil {
		if seconds := in.PublishingIntervalInSeconds; seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return 5 * time.Second
}

// OPCUADevicePropertyType defines the type of property.
// +kubebuilder:validation:Enum=int8;int16;int;int32;int64;uint8;uint16;uint;uint32;uint64;float;float32;double;float64;boolean;string;hexString;binaryString;base64String
type OPCUADevicePropertyType string

const (
	/*
		arithmetic types
	*/

	OPCUADevicePropertyTypeInt8    OPCUADevicePropertyType = "int8"
	OPCUADevicePropertyTypeInt16   OPCUADevicePropertyType = "int16"
	OPCUADevicePropertyTypeInt     OPCUADevicePropertyType = "int" // as same as int32
	OPCUADevicePropertyTypeInt32   OPCUADevicePropertyType = "int32"
	OPCUADevicePropertyTypeInt64   OPCUADevicePropertyType = "int64"
	OPCUADevicePropertyTypeUint8   OPCUADevicePropertyType = "uint8"
	OPCUADevicePropertyTypeUint16  OPCUADevicePropertyType = "uint16"
	OPCUADevicePropertyTypeUint    OPCUADevicePropertyType = "uint" // as same as uint32
	OPCUADevicePropertyTypeUint32  OPCUADevicePropertyType = "uint32"
	OPCUADevicePropertyTypeUint64  OPCUADevicePropertyType = "uint64"
	OPCUADevicePropertyTypeFloat   OPCUADevicePropertyType = "float" // as same as float32
	OPCUADevicePropertyTypeFloat32 OPCUADevicePropertyType = "float32"
	OPCUADevicePropertyTypeDouble  OPCUADevicePropertyType = "double" // as same as float64
	OPCUADevicePropertyTypeFloat64 OPCUADevicePropertyType = "float64"

	/*
		none arithmetic types
	*/

	OPCUADevicePropertyTypeBoolean OPCUADevicePropertyType = "boolean"
	OPCUADevicePropertyTypeString  OPCUADevicePropertyType = "string"

	/*
		for bytes
	*/
	OPCUADevicePropertyTypeHexString    OPCUADevicePropertyType = "hexString"
	OPCUADevicePropertyTypeBinaryString OPCUADevicePropertyType = "binaryString"
	OPCUADevicePropertyTypeBase64String OPCUADevicePropertyType = "base64String"
)

// OPCUADevicePropertyValueArithmeticOperationType defines the type of arithmetic operation.
// +kubebuilder:validation:Enum=Add;Subtract;Multiply;Divide
type OPCUADevicePropertyValueArithmeticOperationType string

const (
	OPCUADevicePropertyValueArithmeticAdd      OPCUADevicePropertyValueArithmeticOperationType = "Add"
	OPCUADevicePropertyValueArithmeticSubtract OPCUADevicePropertyValueArithmeticOperationType = "Subtract"
	OPCUADevicePropertyValueArithmeticMultiply OPCUADevicePropertyValueArithmeticOperationType = "Multiply"
	OPCUADevicePropertyValueArithmeticDivide   OPCUADevicePropertyValueArithmeticOperationType = "Divide"
)

// OPCUADevicePropertyValueArithmeticOperation defines the arithmetic operation of property value.
type OPCUADevicePropertyValueArithmeticOperation struct {
	// Specifies the type of arithmetic operation.
	// +kubebuilder:validation:Required
	Type OPCUADevicePropertyValueArithmeticOperationType `json:"type"`

	// Specifies the value for arithmetic operation, which is in form of float string.
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// OPCUADevicePropertyVisitor defines the visitor of property.
type OPCUADevicePropertyVisitor struct {
	// Specifies the id of OPC-UA node, e.g. "ns=1,i=1005".
	// +kubebuilder:validation:Required
	NodeID string `json:"nodeID"`

	// Specifies the name of OPC-UA node.
	// +optional
	BrowseName string `json:"browseName,omitempty"`

	// Specifies the arithmetic operations in order if needed, only available in arithmetic types.
	// +listType=atomic
	// +optional
	ArithmeticOperations []OPCUADevicePropertyValueArithmeticOperation `json:"arithmeticOperations,omitempty"`

	// Specifies the precision of the arithmetic operation result.
	// The default is "2".
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=24
	// +optional
	ArithmeticOperationPrecision *uint8 `json:"arithmeticOperationPrecision,omitempty"`
}

func (in *OPCUADevicePropertyVisitor) GetArithmeticOperationPrecision() int {
	if in != nil && in.ArithmeticOperationPrecision != nil {
		return int(*in.ArithmeticOperationPrecision)
	}
	return 2
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

	// Specifies the access mode of property.
	// The default value is "ReadMany".
	// +listType=set
	// +kubebuilder:default={ReadMany}
	AccessModes []OPCUADevicePropertyAccessMode `json:"accessModes,omitempty"`

	// Specifies the visitor of property.
	// +kubebuilder:validation:Required
	Visitor OPCUADevicePropertyVisitor `json:"visitor"`

	// Specifies the value of property, only available in the writable property.
	// +optional
	Value string `json:"value,omitempty"`
}

// MergeAccessModes merges the duplicated modes and then returns the access mode array.
func (in *OPCUADeviceProperty) MergeAccessModes() []OPCUADevicePropertyAccessMode {
	if in != nil && len(in.AccessModes) != 0 {
		// NB(thxCode) if both "*Once" and "*Many" are specified,
		// we can merge "*Once" to "*Many" via bitmap,
		// and keep "Write*" before "Read*".
		var mode byte
		for _, accessMode := range in.AccessModes {
			switch accessMode {
			case OPCUADevicePropertyAccessModeNotify:
				mode = mode | 0x10
			case OPCUADevicePropertyAccessModeWriteOnce:
				mode = mode | 0x04
			case OPCUADevicePropertyAccessModeWriteMany:
				mode = mode | 0x08
			case OPCUADevicePropertyAccessModeReadOnce:
				mode = mode | 0x01
			default: // OPCUADevicePropertyAccessModeReadMany
				mode = mode | 0x02
			}
		}
		if mode&0x08 == 0x08 {
			mode = mode & 0xfb
		}
		if mode&0x02 == 0x02 {
			mode = mode & 0xfe
		}

		var accessModes []OPCUADevicePropertyAccessMode
		if mode&0x10 == 0x10 {
			accessModes = append(accessModes, OPCUADevicePropertyAccessModeNotify)
			mode = mode & 0x0f
		}
		switch mode {
		case 0x0a: // 1010
			accessModes = append(accessModes, OPCUADevicePropertyAccessModeWriteMany, OPCUADevicePropertyAccessModeReadMany)
		case 0x09: // 1001
			accessModes = append(accessModes, OPCUADevicePropertyAccessModeWriteMany, OPCUADevicePropertyAccessModeReadOnce)
		case 0x06: // 0110
			accessModes = append(accessModes, OPCUADevicePropertyAccessModeWriteOnce, OPCUADevicePropertyAccessModeReadMany)
		case 0x05: // 0101
			accessModes = append(accessModes, OPCUADevicePropertyAccessModeWriteOnce, OPCUADevicePropertyAccessModeReadOnce)
		case 0x04: // 0100
			accessModes = append(accessModes, OPCUADevicePropertyAccessModeWriteOnce)
		case 0x08: // 1000
			accessModes = append(accessModes, OPCUADevicePropertyAccessModeWriteMany)
		case 0x01: // 0001
			accessModes = append(accessModes, OPCUADevicePropertyAccessModeReadOnce)
		case 0x02: // 0010
			accessModes = append(accessModes, OPCUADevicePropertyAccessModeReadMany)
		}

		// never reach
		return accessModes
	}
	return []OPCUADevicePropertyAccessMode{OPCUADevicePropertyAccessModeReadMany}
}

// OPCUADeviceStatusProperty defines the observed property of OPCUADevice.
type OPCUADeviceStatusProperty struct {
	// Reports the name of property.
	// +optional
	Name string `json:"name,omitempty"`

	// Reports the type of property.
	// +optional
	Type OPCUADevicePropertyType `json:"type,omitempty"`

	// Reports the type of property.
	// +optional
	AccessModes []OPCUADevicePropertyAccessMode `json:"accessModes,omitempty"`

	// Reports the value of property.
	// +optional
	Value string `json:"value,omitempty"`

	// Reports the operation result of property if configured `arithmeticOperations`.
	// +optional
	OperationResult string `json:"operationResult,omitempty"`

	// Reports the updated timestamp of property.
	// +optional
	UpdatedAt *metav1.Time `json:"updatedAt,omitempty"`
}

// OPCUADeviceSpec defines the desired state of OPCUADevice.
type OPCUADeviceSpec struct {
	// Specifies the extension of device.
	// +optional
	Extension *OPCUADeviceExtension `json:"extension,omitempty"`

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

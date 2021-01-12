package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	endianapi "github.com/rancher/octopus/pkg/endian/api"
	mqttapi "github.com/rancher/octopus/pkg/mqtt/api"
)

// MQTTDevicePropertyAccessMode defines the mode for accessing a device property,
// default is "Notify".
// +kubebuilder:validation:Enum=WriteOnce;WriteMany;ReadOnce;Notify
type MQTTDevicePropertyAccessMode string

const (
	MQTTDevicePropertyAccessModeWriteOnce MQTTDevicePropertyAccessMode = "WriteOnce"
	MQTTDevicePropertyAccessModeWriteMany MQTTDevicePropertyAccessMode = "WriteMany"
	MQTTDevicePropertyAccessModeReadOnce  MQTTDevicePropertyAccessMode = "ReadOnce"
	MQTTDevicePropertyAccessModeNotify    MQTTDevicePropertyAccessMode = "Notify" // aka. ReadMany
)

// MQTTDevicePropertyType defines the type of the property value.
// +kubebuilder:validation:Enum=int8;int16;int;int32;int64;uint8;uint16;uint;uint32;uint64;float;float32;double;float64;boolean;hexString;binaryString;base64String;string
type MQTTDevicePropertyType string

const (
	/*
		arithmetic types
	*/

	MQTTDevicePropertyTypeInt8    MQTTDevicePropertyType = "int8"
	MQTTDevicePropertyTypeInt16   MQTTDevicePropertyType = "int16"
	MQTTDevicePropertyTypeInt     MQTTDevicePropertyType = "int" // as same as int32
	MQTTDevicePropertyTypeInt32   MQTTDevicePropertyType = "int32"
	MQTTDevicePropertyTypeInt64   MQTTDevicePropertyType = "int64"
	MQTTDevicePropertyTypeUint8   MQTTDevicePropertyType = "uint8"
	MQTTDevicePropertyTypeUint16  MQTTDevicePropertyType = "uint16"
	MQTTDevicePropertyTypeUint    MQTTDevicePropertyType = "uint" // as same as uint32
	MQTTDevicePropertyTypeUint32  MQTTDevicePropertyType = "uint32"
	MQTTDevicePropertyTypeUint64  MQTTDevicePropertyType = "uint64"
	MQTTDevicePropertyTypeFloat   MQTTDevicePropertyType = "float" // as same as float32
	MQTTDevicePropertyTypeFloat32 MQTTDevicePropertyType = "float32"
	MQTTDevicePropertyTypeDouble  MQTTDevicePropertyType = "double" // as same as float64
	MQTTDevicePropertyTypeFloat64 MQTTDevicePropertyType = "float64"

	/*
		none arithmetic types
	*/

	MQTTDevicePropertyTypeBoolean MQTTDevicePropertyType = "boolean"
	MQTTDevicePropertyTypeString  MQTTDevicePropertyType = "string"

	/*
		for bytes
	*/
	MQTTDevicePropertyTypeHexString    MQTTDevicePropertyType = "hexString"
	MQTTDevicePropertyTypeBinaryString MQTTDevicePropertyType = "binaryString"
	MQTTDevicePropertyTypeBase64String MQTTDevicePropertyType = "base64String"
)

// MQTTDevicePattern defines the pattern that published/subscribed the message.
// AttributedMessage: Compress properties into one message, one topic has its all property values.
// AttributedTopic: Flatten properties to topic, each topic has its own property value.
// +kubebuilder:validation:Enum=AttributedMessage;AttributedTopic
type MQTTDevicePattern string

const (
	MQTTDevicePatternAttributedMessage MQTTDevicePattern = "AttributedMessage"
	MQTTDevicePatternAttributeTopic    MQTTDevicePattern = "AttributedTopic"
)

// MQTTDeviceProtocol is the Schema for configuring the protocol of MQTTDevice.
type MQTTDeviceProtocol struct {
	mqttapi.MQTTOptions `json:",inline"`

	// Specifies the pattern of MQTTDevice protocol.
	// +kubebuilder:validation:Required
	Pattern MQTTDevicePattern `json:"pattern"`

	// Specifies the amount of interval for synchronizing the MQTT device.
	// The default value is "10s".
	// +kubebuilder:default="10s"
	SyncInterval metav1.Duration `json:"syncInterval,omitempty"`
}

func (in *MQTTDeviceProtocol) GetSyncInterval() time.Duration {
	if in != nil {
		if duration := in.SyncInterval.Duration; duration > 0 {
			return duration
		}
	}
	return 10 * time.Second
}

// MQTTDevicePropertyValue defines the value of the property.
// +kubebuilder:validation:Type=""
// +kubebuilder:validation:XPreserveUnknownFields
type MQTTDevicePropertyValue struct {
	Raw []byte `json:"-"`
}

func (in *MQTTDevicePropertyValue) String() string {
	return string(in.Raw)
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (in *MQTTDevicePropertyValue) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && string(data) != "null" {
		in.Raw = data
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (in MQTTDevicePropertyValue) MarshalJSON() ([]byte, error) {
	if len(in.Raw) > 0 {
		return in.Raw, nil
	}
	return nil, nil
}

// ToUnstructured implements the value.UnstructuredConverter interface.
func (in MQTTDevicePropertyValue) ToUnstructured() interface{} {
	if len(in.Raw) > 0 {
		return in.Raw
	}
	return nil
}

// OpenAPISchemaType is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (MQTTDevicePropertyValue) OpenAPISchemaType() []string {
	return nil
}

// OpenAPISchemaFormat is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
func (MQTTDevicePropertyValue) OpenAPISchemaFormat() string { return "" }

// MQTTDevicePropertyValueArithmeticOperationType defines the type of arithmetic operation.
// +kubebuilder:validation:Enum=Add;Subtract;Multiply;Divide
type MQTTDevicePropertyValueArithmeticOperationType string

const (
	MQTTDevicePropertyValueArithmeticAdd      MQTTDevicePropertyValueArithmeticOperationType = "Add"
	MQTTDevicePropertyValueArithmeticSubtract MQTTDevicePropertyValueArithmeticOperationType = "Subtract"
	MQTTDevicePropertyValueArithmeticMultiply MQTTDevicePropertyValueArithmeticOperationType = "Multiply"
	MQTTDevicePropertyValueArithmeticDivide   MQTTDevicePropertyValueArithmeticOperationType = "Divide"
)

// MQTTDevicePropertyValueArithmeticOperation defines the arithmetic operation of property value.
type MQTTDevicePropertyValueArithmeticOperation struct {
	// Specifies the type of arithmetic operation.
	// +kubebuilder:validation:Required
	Type MQTTDevicePropertyValueArithmeticOperationType `json:"type"`

	// Specifies the value for arithmetic operation, which is in form of float string.
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// MQTTDevicePropertyValueContentType defines the content type of property value,
// default is "text".
// +kubebuilder:validation:Enum=text;bytes
type MQTTDevicePropertyValueContentType string

const (
	MQTTDevicePropertyValueContentTypeText  MQTTDevicePropertyValueContentType = "text"
	MQTTDevicePropertyValueContentTypeBytes MQTTDevicePropertyValueContentType = "bytes"
)

// MQTTDevicePropertyVisitor defines the visitor of property.
type MQTTDevicePropertyVisitor struct {
	mqttapi.MQTTMessagePayloadOptions `json:",inline"`
	mqttapi.MQTTMessageTopicOperation `json:",inline"`

	// Specifies the content type of property value.
	// The default is "text".
	// +kubebuilder:default="text"
	ContentType MQTTDevicePropertyValueContentType `json:"contentType,omitempty"`

	// Specifies the endianness of value, only available in basic bytes(content-type) type.
	// The default is "BigEndian".
	// +optional
	Endianness endianapi.DevicePropertyValueEndianness `json:"endianness,omitempty"`

	// Specifies the arithmetic operations in order if needed, only available in arithmetic types.
	// +listType=atomic
	// +optional
	ArithmeticOperations []MQTTDevicePropertyValueArithmeticOperation `json:"arithmeticOperations,omitempty"`

	// Specifies the precision of the arithmetic operation result.
	// The default is "2".
	// +optional
	ArithmeticOperationPrecision *int `json:"arithmeticOperationPrecision,omitempty"`
}

func (in *MQTTDevicePropertyVisitor) GetArithmeticOperationPrecision() int {
	if in != nil && in.ArithmeticOperationPrecision != nil {
		return *in.ArithmeticOperationPrecision
	}
	return 2
}

func (in *MQTTDevicePropertyVisitor) GetEndianness() endianapi.DevicePropertyValueEndianness {
	if in != nil && string(in.Endianness) != "" {
		return in.Endianness
	}
	return endianapi.DevicePropertyValueEndiannessLittleEndian
}

// MQTTDeviceProperty defines the specified property of MQTTDevice.
type MQTTDeviceProperty struct {
	// Specifies the annotations of property.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Specifies the name of property.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Specifies the description of property.
	// +optional
	Description string `json:"description,omitempty"`

	// Specifies the type of property.
	// +kubebuilder:validation:Required
	Type MQTTDevicePropertyType `json:"type"`

	// Specifies the access mode of property.
	// The default value is "Notify".
	// +listType=set
	// +kubebuilder:default={Notify}
	AccessModes []MQTTDevicePropertyAccessMode `json:"accessModes,omitempty"`

	// Specifies the visitor of property.
	// +optional
	Visitor MQTTDevicePropertyVisitor `json:"visitor,omitempty"`

	// Specifies the value of property.
	// +optional
	Value *MQTTDevicePropertyValue `json:"value,omitempty"`
}

// MergeAccessModes merges the duplicated modes and then returns the access mode array.
func (in *MQTTDeviceProperty) MergeAccessModes() []MQTTDevicePropertyAccessMode {
	if in != nil && len(in.AccessModes) != 0 {
		// NB(thxCode) if both "*Once" and "*Many" are specified,
		// we can merge "*Once" to "*Many" via bitmap,
		// and keep "Write*" before "Read*".
		var mode byte
		for _, accessMode := range in.AccessModes {
			switch accessMode {
			case MQTTDevicePropertyAccessModeWriteOnce:
				mode = mode | 0x04
			case MQTTDevicePropertyAccessModeWriteMany:
				mode = mode | 0x08
			case MQTTDevicePropertyAccessModeReadOnce:
				mode = mode | 0x01
			default: // MQTTDevicePropertyAccessModeNotify
				mode = mode | 0x02
			}
		}
		if mode&0x08 == 0x08 {
			mode = mode & 0xfb
		}
		if mode&0x02 == 0x02 {
			mode = mode & 0xfe
		}

		switch mode {
		case 0x0a: // 1010
			return []MQTTDevicePropertyAccessMode{MQTTDevicePropertyAccessModeWriteMany, MQTTDevicePropertyAccessModeNotify}
		case 0x09: // 1001
			return []MQTTDevicePropertyAccessMode{MQTTDevicePropertyAccessModeWriteMany, MQTTDevicePropertyAccessModeReadOnce}
		case 0x06: // 0110
			return []MQTTDevicePropertyAccessMode{MQTTDevicePropertyAccessModeWriteOnce, MQTTDevicePropertyAccessModeNotify}
		case 0x05: // 0101
			return []MQTTDevicePropertyAccessMode{MQTTDevicePropertyAccessModeWriteOnce, MQTTDevicePropertyAccessModeReadOnce}
		case 0x04: // 0100
			return []MQTTDevicePropertyAccessMode{MQTTDevicePropertyAccessModeWriteOnce}
		case 0x08: // 1000
			return []MQTTDevicePropertyAccessMode{MQTTDevicePropertyAccessModeWriteMany}
		case 0x01: // 0001
			return []MQTTDevicePropertyAccessMode{MQTTDevicePropertyAccessModeReadOnce}
		default: // 0010
		}
	}
	return []MQTTDevicePropertyAccessMode{MQTTDevicePropertyAccessModeNotify}
}

// MQTTDeviceStatusProperty defines the observed property of MQTTDevice.
type MQTTDeviceStatusProperty struct {
	// Reports the name of property.
	// +optional
	Name string `json:"name,omitempty"`

	// Reports the type of property.
	// +optional
	Type MQTTDevicePropertyType `json:"type,omitempty"`

	// Reports the access mode of property.
	// +optional
	AccessModes []MQTTDevicePropertyAccessMode `json:"accessModes,omitempty"`

	// Reports the value of property.
	// +optional
	Value string `json:"value,omitempty"`

	// Reports the operation result of property if configured `arithmeticOperations`.
	// +optional
	OperationResult string `json:"operationResult,omitempty"`

	// Reports the updated timestamp of property.
	// +optional
	UpdatedAt *metav1.Time `json:"updateAt,omitempty"`
}

// MQTTDeviceSpec defines the desired state of MQTTDevice.
type MQTTDeviceSpec struct {
	// Specifies the protocol for accessing the MQTT service.
	// +kubebuilder:validation:Required
	Protocol MQTTDeviceProtocol `json:"protocol"`

	// Specifies the properties of MQTTDevice.
	// +listType=map
	// +listMapKey=name
	// +optional
	Properties []MQTTDeviceProperty `json:"properties,omitempty"`
}

// MQTTDeviceStatus defines the observed state of MQTTDevice.
type MQTTDeviceStatus struct {
	// Reports the properties of MQTTDevice.
	// +optional
	Properties []MQTTDeviceStatusProperty `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=mqtt
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PATTERN",type="string",JSONPath=`.spec.protocol.pattern`
// +kubebuilder:printcolumn:name="SERVER",type="string",JSONPath=`.spec.protocol.client.server`
// +kubebuilder:printcolumn:name="TOPIC",type="string",JSONPath=`.spec.protocol.message.topic`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=`.metadata.creationTimestamp`
// MQTTDevice is the Schema for the MQTT device API.
type MQTTDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MQTTDeviceSpec   `json:"spec,omitempty"`
	Status MQTTDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// MQTTDeviceList contains a list of MQTTDevice.
type MQTTDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MQTTDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MQTTDevice{}, &MQTTDeviceList{})
}

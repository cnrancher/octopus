package v1alpha1

import (
	"fmt"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// Device Property value type
	ValueTypeInt     ValueType = "int"
	ValueTypeString  ValueType = "string"
	ValueTypeFloat   ValueType = "float"
	ValueTypeBoolean ValueType = "boolean"
	ValueTypeArray   ValueType = "array"
	ValueTypeObject  ValueType = "object"

	// Subscribed topic payload type
	PayloadTypeJSON PayloadType = "json"
)

// Defines the type of the property value.
// +kubebuilder:validation:Enum=int;string;float;boolean;array;object
type ValueType string

// The payload type type.
// +kubebuilder:validation:Enum=json
type PayloadType string

// The qos type.
// +kubebuilder:validation:Enum=0;1;2
type QosType int

type MqttConfig struct {
	Broker   string `json:"broker"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type SubInfo struct {
	Topic       string      `json:"topic"`
	PayloadType PayloadType `json:"payloadType"`
	Qos         QosType     `json:"qos"`
}

type PubInfo struct {
	Topic string  `json:"topic"`
	Qos   QosType `json:"qos"`
}

type ValueFloat struct {
	F float64 `json:"-"`
}

func (v *ValueFloat) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`"%f"`, v.F)
	return []byte(str), nil
}
func (v *ValueFloat) UnmarshalJSON(value []byte) error {
	var err error
	s := value
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	v.F, err = strconv.ParseFloat(string(s), 64)
	return err
}

type ValueArrayProps struct {
	// +kubebuilder:validation:XPreserveUnknownFields
	ValueProps `json:",inline"`
}

type ValueProps struct {
	// Reports the type of property.
	ValueType ValueType `json:"valueType"`

	// Reports the value of int type.
	// +optional
	IntValue int64 `json:"intValue,omitempty"`

	// Reports the value of string type.
	// +optional
	StringValue string `json:"stringValue,omitempty"`

	// Reports the value of float type.
	// +optional
	FloatValue *ValueFloat `json:"floatValue,omitempty"`

	// Reports the value of boolean type.
	// +optional
	BooleanValue bool `json:"booleanValue,omitempty"`

	// Reports the value of array type.
	// +optional
	ArrayValue []ValueArrayProps `json:"arrayValue,omitempty"`

	// Reports the value of object type.
	// +kubebuilder:validation:XPreserveUnknownFields
	// +optional
	ObjectValue *runtime.RawExtension `json:"objectValue,omitempty"`
}

type Property struct {
	SubInfo     SubInfo    `json:"subInfo"`
	PubInfo     PubInfo    `json:"pubInfo,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	JSONPath    string     `json:"jsonPath"`
	Value       ValueProps `json:"value,omitempty"`
}

type StatusProperty struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Value       ValueProps  `json:"value"`
	UpdatedAt   metav1.Time `json:"updateAt"`
}

// MqttDeviceSpec defines the desired state of MqttDevice
type MqttDeviceSpec struct {
	Config     MqttConfig `json:"config"`
	Properties []Property `json:"properties"`
}

// MqttDeviceStatus defines the observed state of MqttDevice
type MqttDeviceStatus struct {
	Properties []StatusProperty `json:"properties"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// MqttDevice is the Schema for the mqtt device API
type MqttDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MqttDeviceSpec `json:"spec,omitempty"`
	// +kubebuilder:validation:XPreserveUnknownFields
	Status MqttDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// MqttDeviceList contains a list of MqttDevice
type MqttDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MqttDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MqttDevice{}, &MqttDeviceList{})
}

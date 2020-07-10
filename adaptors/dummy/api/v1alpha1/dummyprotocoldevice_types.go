package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DummyProtocolDevicePropertyType defines the type of property.
// +kubebuilder:validation:Enum=string;int;float;boolean;array;object
type DummyProtocolDevicePropertyType string

const (
	DummyProtocolDevicePropertyTypeInt     DummyProtocolDevicePropertyType = "int"
	DummyProtocolDevicePropertyTypeFloat   DummyProtocolDevicePropertyType = "float"
	DummyProtocolDevicePropertyTypeString  DummyProtocolDevicePropertyType = "string"
	DummyProtocolDevicePropertyTypeBoolean DummyProtocolDevicePropertyType = "boolean"
	DummyProtocolDevicePropertyTypeArray   DummyProtocolDevicePropertyType = "array"
	DummyProtocolDevicePropertyTypeObject  DummyProtocolDevicePropertyType = "object"
)

type DummyProtocolDeviceObjectOrArrayProperty struct {
	// +kubebuilder:validation:XPreserveUnknownFields
	DummyProtocolDeviceProperty `json:",inline"`
}

// DummyProtocolDeviceProperty defines the desired property of DummyProtocolDevice.
type DummyProtocolDeviceProperty struct {
	// Specifies the description of property.
	// +optional
	Description string `json:"description,omitempty"`

	// Specifies the type of property.
	// +kubebuilder:validation:Required
	Type DummyProtocolDevicePropertyType `json:"type"`

	// Specifies if the property is readonly.
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`

	// Specifies the item property if the type is "array".
	// +optional
	ArrayProperties *DummyProtocolDeviceObjectOrArrayProperty `json:"arrayProperties,omitempty"`

	// Specifies the object property if the type is "object".
	// +optional
	ObjectProperties map[string]DummyProtocolDeviceObjectOrArrayProperty `json:"objectProperties,omitempty"`
}

// DummyProtocolDeviceProtocol defines the desired protocol of DummyProtocolDevice.
type DummyProtocolDeviceProtocol struct {
	// Specifies the IP address of device.
	// +kubebuilder:validation:Required
	IP string `json:"ip"`
}

// DummyProtocolDeviceSpec defines the desired state of DummyProtocolDevice.
type DummyProtocolDeviceSpec struct {
	// Specifies the extension of device.
	// +optional
	Extension *DummyDeviceExtension `json:"extension,omitempty"`

	// Specifies the protocol for accessing the device.
	// +kubebuilder:validation:Required
	Protocol DummyProtocolDeviceProtocol `json:"protocol"`

	// Specifies the properties of device.
	// +optional
	Properties map[string]DummyProtocolDeviceProperty `json:"properties,omitempty"`
}

type DummyProtocolDeviceStatusObjectOrArrayProperty struct {
	// +kubebuilder:validation:XPreserveUnknownFields
	DummyProtocolDeviceStatusProperty `json:",inline"`
}

// DummyProtocolDeviceStatusProperty defines the observed property of DummyProtocolDevice.
type DummyProtocolDeviceStatusProperty struct {
	// Reports the type of property.
	// +optional
	Type DummyProtocolDevicePropertyType `json:"type,omitempty"`

	// Reports the value of int type.
	// +optional
	IntValue *int `json:"intValue,omitempty"`

	// Reports the value of string type.
	// +optional
	StringValue *string `json:"stringValue,omitempty"`

	// Reports the value of float type.
	// +optional
	FloatValue *resource.Quantity `json:"floatValue,omitempty"`

	// Reports the value of boolean type.
	// +optional
	BooleanValue *bool `json:"booleanValue,omitempty"`

	// Reports the value of array type.
	// +optional
	ArrayValue []DummyProtocolDeviceStatusObjectOrArrayProperty `json:"arrayValue,omitempty"`

	// Reports the value of object type.
	// +optional
	ObjectValue map[string]DummyProtocolDeviceStatusObjectOrArrayProperty `json:"objectValue,omitempty"`
}

// DummyProtocolDeviceStatus defines the observed state of DummyProtocolDevice.
type DummyProtocolDeviceStatus struct {
	// Reports the properties of device.
	// +optional
	Properties map[string]DummyProtocolDeviceStatusProperty `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=dummyprotocol
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="IP",type="string",JSONPath=`.spec.protocol.ip`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=`.metadata.creationTimestamp`
// DummyProtocolDevice is the schema for the dummy protocol device API.
type DummyProtocolDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DummyProtocolDeviceSpec   `json:"spec,omitempty"`
	Status DummyProtocolDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// DummyProtocolDeviceList contains a list of DummyProtocolDevice.
type DummyProtocolDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DummyProtocolDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DummyProtocolDevice{}, &DummyProtocolDeviceList{})
}

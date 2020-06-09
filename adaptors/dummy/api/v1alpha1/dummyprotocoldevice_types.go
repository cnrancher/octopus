package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DummyProtocolDevicePropertyType describes the type of property.
// +kubebuilder:validation:Enum=string;int;float;boolean;array;object
type DummyProtocolDevicePropertyType string

const (
	DummyProtocolDevicePropertyTypeString  DummyProtocolDevicePropertyType = "string"
	DummyProtocolDevicePropertyTypeInt     DummyProtocolDevicePropertyType = "int"
	DummyProtocolDevicePropertyTypeFloat   DummyProtocolDevicePropertyType = "float"
	DummyProtocolDevicePropertyTypeBoolean DummyProtocolDevicePropertyType = "boolean"
	DummyProtocolDevicePropertyTypeArray   DummyProtocolDevicePropertyType = "array"
	DummyProtocolDevicePropertyTypeObject  DummyProtocolDevicePropertyType = "object"
)

type DummyProtocolDeviceSpecObjectOrArrayProps struct {
	// +kubebuilder:validation:XPreserveUnknownFields
	DummyProtocolDeviceSpecProps `json:",inline"`
}

// DummyProtocolDeviceSpecProps defines the property of DummyProtocolDeviceSpec.
type DummyProtocolDeviceSpecProps struct {
	// Describes the type of property.
	// +kubebuilder:validation:Required
	Type DummyProtocolDevicePropertyType `json:"type"`

	// Outlines the property.
	// +optional
	Description string `json:"description,omitempty"`

	// Configures the property is readOnly or not.
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`

	// Describes item properties of the array type.
	// +optional
	ArrayProps *DummyProtocolDeviceSpecObjectOrArrayProps `json:"arrayProps,omitempty"`

	// Describes properties of the object type.
	// +optional
	ObjectProps map[string]DummyProtocolDeviceSpecObjectOrArrayProps `json:"objectProps,omitempty"`
}

// DummyProtocolDeviceProtocol describes the accessing protocol for dummy protocol device.
type DummyProtocolDeviceProtocol struct {
	// Specifies where to connect the dummy protocol device.
	IP string `json:"ip"`
}

// DummyProtocolDeviceSpec defines the desired state of DummyProtocolDevice.
type DummyProtocolDeviceSpec struct {
	// Specifies the extension of device.
	// +optional
	Extension DeviceExtensionSpec `json:"extension,omitempty"`

	// Protocol for accessing the dummy protocol device.
	// +kubebuilder:validation:Required
	Protocol DummyProtocolDeviceProtocol `json:"protocol"`

	// Describe the desired properties.
	// +optional
	Props map[string]DummyProtocolDeviceSpecProps `json:"props,omitempty"`
}

type DummyProtocolDeviceStatusObjectOrArrayProps struct {
	// +kubebuilder:validation:XPreserveUnknownFields
	DummyProtocolDeviceStatusProps `json:",inline"`
}

// DummyProtocolDeviceStatusProps defines the property of DummyProtocolDeviceStatus.
type DummyProtocolDeviceStatusProps struct {
	// Reports the type of property.
	Type DummyProtocolDevicePropertyType `json:"type"`

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
	ArrayValue []DummyProtocolDeviceStatusObjectOrArrayProps `json:"arrayValue,omitempty"`

	// Reports the value of object type.
	// +optional
	ObjectValue map[string]DummyProtocolDeviceStatusObjectOrArrayProps `json:"objectValue,omitempty"`
}

// DummyProtocolDeviceStatus defines the observed state of DummyProtocolDevice.
type DummyProtocolDeviceStatus struct {
	// Reports the extension of device.
	// +optional
	Extension *DeviceExtensionStatus `json:"extension,omitempty"`

	// Reports the observed value of the desired properties.
	// +optional
	Props map[string]DummyProtocolDeviceStatusProps `json:"props,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// DummyProtocolDevice is the Schema for the dummy protocol device API.
type DummyProtocolDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DummyProtocolDeviceSpec   `json:"spec,omitempty"`
	Status DummyProtocolDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// DummyProtocolDeviceList contains a list of DummyProtocolDevice
type DummyProtocolDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DummyProtocolDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DummyProtocolDevice{}, &DummyProtocolDeviceList{})
}

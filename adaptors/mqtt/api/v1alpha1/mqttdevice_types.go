package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mqttapi "github.com/rancher/octopus/pkg/mqtt/api"
)

// MQTTDevicePropertyType defines the type of the property value.
// +kubebuilder:validation:Enum=int;string;float;boolean;object;array
type MQTTDevicePropertyType string

const (
	MQTTDevicePropertyTypeInt     MQTTDevicePropertyType = "int"
	MQTTDevicePropertyTypeString  MQTTDevicePropertyType = "string"
	MQTTDevicePropertyTypeFloat   MQTTDevicePropertyType = "float"
	MQTTDevicePropertyTypeBoolean MQTTDevicePropertyType = "boolean"
	MQTTDevicePropertyTypeObject  MQTTDevicePropertyType = "object"
	MQTTDevicePropertyTypeArray   MQTTDevicePropertyType = "array"
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

// MQTTDeviceSchema defines the pattern schema.
type MQTTDeviceSchema struct {
	// Specifies the type of schema.
	// +optional
	Type string `json:"type,omitempty"`

	// Specifies the reference for schema.
	// +optional
	Reference string `json:"reference,omitempty"`
}

// MQTTDeviceProtocol is the Schema for configuring the protocol of MQTTDevice.
type MQTTDeviceProtocol struct {
	mqttapi.MQTTOptions `json:",inline"`

	// Specifies the pattern of MQTTDevice protocol.
	// +kubebuilder:validation:Required
	Pattern MQTTDevicePattern `json:"pattern"`

	// Specifies the schema of the pattern.
	// +optional
	Schema *MQTTDeviceSchema `json:"schema,omitempty"`
}

// MQTTDevicePropertyValue defines the value of the property.
// +kubebuilder:validation:Type=""
// +kubebuilder:validation:XPreserveUnknownFields
type MQTTDevicePropertyValue struct {
	Raw []byte `json:"-"`
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

// OpenAPISchemaType is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (MQTTDevicePropertyValue) OpenAPISchemaType() []string {
	// TODO: return actual types when anyOf is supported
	return nil
}

// OpenAPISchemaFormat is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
func (MQTTDevicePropertyValue) OpenAPISchemaFormat() string { return "" }

// MQTTDeviceProperty defines the specified property of MQTTDevice.
type MQTTDeviceProperty struct {
	mqttapi.MQTTMessagePayloadOptions `json:",inline"`
	mqttapi.MQTTMessageTopicOperation `json:",inline"`

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
	Type MQTTDevicePropertyType `json:"type,omitempty"`

	// Specifies the value of property.
	// +optional
	Value *MQTTDevicePropertyValue `json:"value,omitempty"`

	// Specifies the MIME of property value.
	// +optional
	ContentType string `json:"contentType,omitempty"`

	// Specifies if the property is read-only.
	// The default value is "true".
	// +kubebuilder:default=true
	ReadOnly *bool `json:"readOnly,omitempty"`
}

// MQTTDeviceStatusProperty defines the observed property of MQTTDevice.
type MQTTDeviceStatusProperty struct {
	MQTTDeviceProperty `json:",inline"`

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

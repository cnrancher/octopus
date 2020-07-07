package api

import (
	"github.com/rancher/octopus/pkg/util/converter"
)

// MQTTMessageQoSLevel defines the QoS level of publishing message.
// 0: Send at most once.
// 1: Send at least once, default.
// 2: Send exactly once.
// +kubebuilder:validation:Enum=0;1;2
type MQTTMessageQoSLevel byte

// MQTTMessageTopicOperator defines the operator for rendering the topic.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTMessageTopicOperator struct {
	// Specifies the operator for rendering the `:operator` keyword of topic during subscribing.
	// +optional
	Read string `json:"read,omitempty"`

	// Specifies the operator for rendering the `:operator` keyword of topic during publishing.
	// +optional
	Write string `json:"write,omitempty"`
}

// MQTTMessageTopicOperation defines the operation of MQTT topic.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTMessageTopicOperation struct {
	// Specifies the path for rendering the `:path` keyword of topic.
	// +optional
	Path string `json:"path,omitempty"`

	// Specifies the operator for rendering the `:operator` keyword of topic.
	// +optional
	Operator *MQTTMessageTopicOperator `json:"operator,omitempty"`
}

// MQTTMessagePayloadOptions defines the options of MQTT message payload.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTMessagePayloadOptions struct {
	// Specifies the QoS of the message.
	// The default value is "1".
	// +kubebuilder:default=1
	// +optional
	QoS *MQTTMessageQoSLevel `json:"qos,omitempty"`

	// Specifies if the last published message to be retained.
	// The default value is "true".
	// +kubebuilder:default=true
	// +optional
	Retained *bool `json:"retained,omitempty"`
}

// MQTTWillMessageContent defines the content of will message.
// +kubebuilder:validation:Type=string
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTWillMessageContent struct {
	Data []byte `json:"-"`
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (in *MQTTWillMessageContent) UnmarshalJSON(b []byte) error {
	// NB(thxCode) refer to https://github.com/kubernetes/apimachinery/blob/3c2682fedbf25e98d89c710d5d707636ac0d6ce4/pkg/util/intstr/intstr.go#L83-L86
	// we can simply remove the double quotes.
	var raw, err = converter.DecodeBase64(b[1 : len(b)-1])
	if err != nil {
		return err
	}

	in.Data = raw
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (in MQTTWillMessageContent) MarshalJSON() ([]byte, error) {
	return converter.EncodeBase64(in.Data), nil
}

// ToUnstructured implements the value.UnstructuredConverter interface.
func (in MQTTWillMessageContent) ToUnstructured() interface{} {
	return converter.UnsafeBytesToString(converter.EncodeBase64(in.Data))
}

// OpenAPISchemaType is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (MQTTWillMessageContent) OpenAPISchemaType() []string { return []string{"string"} }

// OpenAPISchemaFormat is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
func (MQTTWillMessageContent) OpenAPISchemaFormat() string { return "" }

// MQTTWillMessage defines the will message of MQTT client.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTWillMessage struct {
	// Specifies the topic of will message.
	// if not set, the topic will append "$will" to the topic name specified
	// in parent field as its topic name.
	// +kubebuilder:validation:Pattern=".*[^/]$"
	// +optional
	Topic string `json:"topic,omitempty"`

	// Specifies the content of will message. The serialized form of the content is a
	// base64 encoded string, representing the arbitrary (possibly non-string) content value here.
	// +kubebuilder:validation:Required
	Content MQTTWillMessageContent `json:"content"`
}

// MQTTMessageOptions defines the options of MQTT message.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTMessageOptions struct {
	MQTTMessagePayloadOptions `json:",inline"`
	MQTTMessageTopicOperation `json:",inline"`

	// Specifies the topic.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=".*[^/]$"
	Topic string `json:"topic"`

	// Specifies the will message.
	// +optional
	Will *MQTTWillMessage `json:"will,omitempty"`
}

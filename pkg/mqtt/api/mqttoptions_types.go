package api

// MQTTOptions defines the desired state of MQTT client.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type MQTTOptions struct {
	// Specifies the client settings.
	// +kubebuilder:validation:Required
	Client MQTTClientOptions `json:"client"`

	// Specifies the message settings.
	// +kubebuilder:validation:Required
	Message MQTTMessageOptions `json:"message"`
}

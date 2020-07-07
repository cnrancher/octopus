package v1alpha1

import mqttapi "github.com/rancher/octopus/pkg/mqtt/api"

// DeviceExtensionSpec defines the desired state of device extension.
type DeviceExtensionSpec struct {
	// Specifies the MQTT settings.
	// +optional
	MQTT *mqttapi.MQTTOptions `json:"mqtt,omitempty"`
}

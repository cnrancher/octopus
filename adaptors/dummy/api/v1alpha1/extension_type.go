package v1alpha1

import "github.com/rancher/octopus/pkg/mqtt/api/v1alpha1"

// DeviceExtensionSpec defines the desired state of device extension.
type DeviceExtensionSpec struct {
	// Specifies the MQTT settings.
	// +optional
	MQTT *v1alpha1.MQTTOptionsSpec `json:"mqtt,omitempty"`
}

// DeviceExtensionStatus defines the observed state of device extension.
type DeviceExtensionStatus struct {
	// Reports the MQTT settings.
	// +optional
	MQTT *v1alpha1.MQTTOptionsStatus `json:"mqtt,omitempty"`
}

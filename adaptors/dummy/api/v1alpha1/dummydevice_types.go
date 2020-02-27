package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DummyDeviceGear describes how fast the dummy device should be.
// +kubebuilder:validation:Enum=slow;middle;fast
type DummyDeviceGear string

const (
	Slow   DummyDeviceGear = "slow"
	Middle DummyDeviceGear = "middle"
	Fast   DummyDeviceGear = "fast"
)

// DummyDeviceSpec defines the desired state of DummyDevice
type DummyDeviceSpec struct {
	// Turn on the device
	// +kubebuilder:validation:Required
	On bool `json:"on"`

	// Specifies how fast the device should be
	// +optional
	Gear DummyDeviceGear `json:"gear,omitempty"`
}

// DummyDeviceStatus defines the observed state of DummyDevice
type DummyDeviceStatus struct {
	// Reports the current gear
	// +optional
	Gear DummyDeviceGear `json:"gear,omitempty"`

	// Reports the detail number of speed
	// +optional
	RotatingSpeed int32 `json:"rotatingSpeed,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="GEAR",type=string,JSONPath=`.status.gear`
// +kubebuilder:printcolumn:name="SPEED",type=integer,JSONPath=`.status.rotatingSpeed`,format=int32
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// DummyDevice is the Schema for the dummy device API
type DummyDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DummyDeviceSpec   `json:"spec,omitempty"`
	Status DummyDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// DummyDeviceList contains a list of Fan
type DummyDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DummyDevice `json:"items"`
}

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DummySpecialDeviceProtocol describes the accessing protocol for dummy special device.
type DummySpecialDeviceProtocol struct {
	// Specifies where to locate the dummy special device.
	Location string `json:"location"`
}

// DummySpecialDeviceGear describes how fast the dummy special device should be.
// +kubebuilder:validation:Enum=slow;middle;fast
type DummySpecialDeviceGear string

const (
	DummySpecialDeviceGearSlow   DummySpecialDeviceGear = "slow"
	DummySpecialDeviceGearMiddle DummySpecialDeviceGear = "middle"
	DummySpecialDeviceGearFast   DummySpecialDeviceGear = "fast"
)

// DummySpecialDeviceSpec defines the desired state of DummySpecialDevice.
type DummySpecialDeviceSpec struct {
	// Specifies the extension of device.
	// +optional
	Extension DeviceExtensionSpec `json:"extension,omitempty"`

	// Protocol for accessing the dummy special device.
	// +kubebuilder:validation:Required
	Protocol DummySpecialDeviceProtocol `json:"protocol"`

	// Turn on the dummy special device.
	// +kubebuilder:validation:Required
	On bool `json:"on"`

	// Specifies how fast the dummy special device should be.
	// +optional
	Gear DummySpecialDeviceGear `json:"gear,omitempty"`
}

// DummySpecialDeviceStatus defines the observed state of DummySpecialDevice.
type DummySpecialDeviceStatus struct {
	// Reports the extension of device.
	// +optional
	Extension *DeviceExtensionStatus `json:"extension,omitempty"`

	// Reports the current gear of dummy special device.
	// +optional
	Gear DummySpecialDeviceGear `json:"gear,omitempty"`

	// Reports the detail number of speed of dummy special device.
	// +optional
	RotatingSpeed int32 `json:"rotatingSpeed,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="GEAR",type=string,JSONPath=`.status.gear`
// +kubebuilder:printcolumn:name="SPEED",type=integer,JSONPath=`.status.rotatingSpeed`,format=int32
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// DummySpecialDevice is the Schema for the dummy special device API.
type DummySpecialDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DummySpecialDeviceSpec   `json:"spec,omitempty"`
	Status DummySpecialDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// DummySpecialDeviceList contains a list of DummySpecialDevice.
type DummySpecialDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DummySpecialDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DummySpecialDevice{}, &DummySpecialDeviceList{})
}

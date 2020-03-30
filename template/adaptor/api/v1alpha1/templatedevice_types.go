package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TemplateDeviceSpec defines the desired state of TemplateDevice
type TemplateDeviceSpec struct {
	// TODO to fill
}

// TemplateDeviceStatus defines the observed state of TemplateDevice
type TemplateDeviceStatus struct {
	// TODO to fill
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// TemplateDevice is the Schema for the template device API
type TemplateDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateDeviceSpec   `json:"spec,omitempty"`
	Status TemplateDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// TemplateDeviceList contains a list of TemplateDevice
type TemplateDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TemplateDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TemplateDevice{}, &TemplateDeviceList{})
}

package v1alpha1

import (
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgentDeviceGroupSpec defines the desired state of AgentDeviceGroup
type AgentDeviceGroupSpec struct {
	// Describe the app that will be created.
	Apps []App `json:"apps,omitempty"`
	// Use specified endpoint for registration instead of kubernetes cluster endpoint
	ServerURL string `json:"serverURL,omitempty"`
	// Whether to delete the added nodes or not when delete this AgentDeviceGroup
	DeleteNodes bool `json:"deleteNodes,omitempty"`
}

type App struct {
	// DaemonSetStatus name
	Name      string      `json:"name"`
	Namespace string      `json:"namespace"`
	Template  PodTemplate `json:"template"`
}

type AppStatus struct {
	Name            string             `json:"name"`
	Namespace       string             `json:"namespace"`
	DaemonSetStatus v1.DaemonSetStatus `json:"daemonSetStatus"`
	UpdatedAt       metav1.Time        `json:"updatedAt,omitempty"`
}

type PodTemplate struct {
	PodMetadata `json:"metadata,omitempty"`
	Spec        corev1.PodSpec `json:"spec,omitempty"`
}

type PodMetadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}

// AgentDeviceGroupStatus defines the observed state of AgentDeviceGroup
type AgentDeviceGroupStatus struct {
	RegisterCommand string      `json:"command,omitempty"`
	Nodes           []string    `json:"nodes,omitempty"`
	Apps            []AppStatus `json:"apps,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// AgentDeviceGroup is the Schema for the AgentDeviceGroup API
type AgentDeviceGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentDeviceGroupSpec   `json:"spec,omitempty"`
	Status AgentDeviceGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// AgentDeviceGroupList contains a list of AgentDeviceGroup
type AgentDeviceGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentDeviceGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgentDeviceGroup{}, &AgentDeviceGroupList{})
}

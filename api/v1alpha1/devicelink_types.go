package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeviceAdaptor defines the properties of device adaptor
type DeviceAdaptor struct {
	// Specifies the node of adaptor to be matched.
	// +optional
	Node string `json:"node,omitempty"`

	// Specifies the name of adaptor to be used.
	// +optional
	Name string `json:"name,omitempty"`

	// Specifies the parameter of adaptor to be used.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`
}

// DeviceMeta defines the metadata of device
type DeviceMeta struct {
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// DeviceTemplateSpec defines the device desired state
type DeviceTemplateSpec struct {
	// Standard object's metadata.
	// +optional
	DeviceMeta `json:"metadata,omitempty"`

	// Specifies the desired behaviors of a device.
	// +kubebuilder:validation:XPreserveUnknownFields
	// +optional
	Spec *runtime.RawExtension `json:"spec,omitempty"`
}

type DeviceLinkConditionType string

// These are valid conditions of a device
const (
	// NodeExisted means that if the node was existed,
	// it is ready for validating model.
	DeviceLinkNodeExisted DeviceLinkConditionType = "NodeExisted"

	// ModelExisted means that if the CRD of the model was existed,
	// it is ready for validating adaptor.
	DeviceLinkModelExisted DeviceLinkConditionType = "ModelExisted"

	// AdaptorExisted means that if the adaptor was existed,
	// it is ready for validating device.
	DeviceLinkAdaptorExisted DeviceLinkConditionType = "AdaptorExisted"

	// DeviceCreated means that if the device was created,
	// it is ready for connecting device.
	DeviceLinkDeviceCreated DeviceLinkConditionType = "DeviceCreated"

	// DeviceConnected means the connection of device is healthy.
	DeviceLinkDeviceConnected DeviceLinkConditionType = "DeviceConnected"
)

// DeviceLinkCondition describes the state of a device at a certain point.
type DeviceLinkCondition struct {
	// Type of device condition.
	Type DeviceLinkConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	Status metav1.ConditionStatus `json:"status"`

	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// DeviceLinkSpec defines the desired state of DeviceLink
type DeviceLinkSpec struct {
	// Specifies the desired adaptor of a device
	// +kubebuilder:validation:Required
	Adaptor DeviceAdaptor `json:"adaptor"`

	// Specifies the desired model of a device.
	// +kubebuilder:validation:Required
	Model metav1.TypeMeta `json:"model"`

	// Describe the device that will be created.
	// +kubebuilder:validation:Required
	Template DeviceTemplateSpec `json:"template"`
}

// DeviceLinkStatus defines the observed state of DeviceLink
type DeviceLinkStatus struct {
	// Represents the latest available observations of the device's current state.
	// +optional
	Conditions []DeviceLinkCondition `json:"conditions,omitempty"`

	// Represents the observed scheduled Node name of the device.
	// +optional
	NodeName string `json:"nodeName,omitempty"`

	// Represents the observed scheduled Node hostname of the device.
	// +optional
	NodeHostName string `json:"nodeHostName,omitempty"`

	// Represents the observed scheduled Node internal IP of the device.
	// +optional
	NodeInternalIP string `json:"nodeInternalIP,omitempty"`

	// Represents the observed scheduled Node internal DNS of the device.
	// +optional
	NodeInternalDNS string `json:"nodeInternalDNS,omitempty"`

	// Represents the observed scheduled Node external IP of the device.
	// +optional
	NodeExternalIP string `json:"nodeExternalIP,omitempty"`

	// Represents the observed scheduled Node external DNS of the device.
	// +optional
	NodeExternalDNS string `json:"nodeExternalDNS,omitempty"`

	// Represents the observed model of the device.
	// +optional
	Model metav1.TypeMeta `json:"model,omitempty"`

	// Represents the observed adaptor name of the device.
	// +optional
	AdaptorName string `json:"adaptorName,omitempty"`

	// Represents the observed template generation of the device.
	// +optional
	DeviceTemplateGeneration int64 `json:"deviceTemplateGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=dl
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="KIND",type=string,JSONPath=`.spec.model.kind`
// +kubebuilder:printcolumn:name="NODE",type=string,JSONPath=`.spec.adaptor.node`
// +kubebuilder:printcolumn:name="ADAPTOR",type=string,JSONPath=`.spec.adaptor.name`
// +kubebuilder:printcolumn:name="PHASE",type=string,JSONPath=`.status.conditions[-1].type`
// +kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.conditions[-1].reason`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// DeviceLink is the Schema for the devicelinks API
type DeviceLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceLinkSpec   `json:"spec,omitempty"`
	Status DeviceLinkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// DeviceLinkList contains a list of DeviceLink
type DeviceLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeviceLink `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeviceLink{}, &DeviceLinkList{})
}

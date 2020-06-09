package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeviceLinkReferenceRelationship defines the relationship to refer the reference item of DeviceLink.
// +kubebuilder:object:generate=true
// +kubebuilder:object:root=false
type DeviceLinkReferenceRelationship struct {
	// Specifies the name of reference.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Specifies the item name of the referred reference.
	// +kubebuilder:validation:Required
	Item string `json:"item"`
}

// DeviceLinkReferenceSecretSource defines the source of a same name Secret instance.
type DeviceLinkReferenceSecretSource struct {
	// Specifies the name of the Secret in the same Namespace to use.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Specifies the key of the Secret's data.
	// If not specified, all keys of the Secret will be projected into the parameter values.
	// If specified, the listed keys will be projected into the parameter value.
	// If a key is specified which is not present in the Secret,
	// the connection will error unless it is marked optional.
	// +optional
	Items []string `json:"items,omitempty"`
}

// DeviceLinkReferenceConfigMapSource defines the source of a same name ConfigMap instance.
type DeviceLinkReferenceConfigMapSource struct {
	// Specifies the name of the ConfigMap in the same Namespace to use.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Specifies the key of the ConfigMap's data.
	// If not specified, all keys of the ConfigMap will be projected into the parameter values.
	// If specified, the listed keys will be projected into the parameter value.
	// If a key is specified which is not present in the ConfigMap,
	// the connection will error unless it is marked optional.
	// +optional
	Items []string `json:"items,omitempty"`
}

// DeviceLinkReferenceDownwardAPISourceItem defines the downward API item for projecting the DeviceLink.
type DeviceLinkReferenceDownwardAPISourceItem struct {
	// Specifies the key of the downward API's data.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Specifies that how to select a field of the DeviceLink,
	// only annotations, labels, name, namespace and status are supported.
	// +kubebuilder:validation:Required
	FieldRef *corev1.ObjectFieldSelector `json:"fieldRef"`
}

// DeviceLinkReferenceDownwardAPISource defines the downward API for projecting the DeviceLink.
type DeviceLinkReferenceDownwardAPISource struct {
	// Specifies a list of downward API.
	// +kubebuilder:validation:MinItems=1
	Items []DeviceLinkReferenceDownwardAPISourceItem `json:"items"`
}

// DeviceLinkReferenceSource defines the parameter source.
type DeviceLinkReferenceSource struct {
	// Secret represents a Secret of the same Namespace that should populate this connection.
	// +optional
	Secret *DeviceLinkReferenceSecretSource `json:"secret,omitempty"`

	// ConfigMap represents a ConfigMap of the same Namespace that should populate this connection.
	// +optional
	ConfigMap *DeviceLinkReferenceConfigMapSource `json:"configMap,omitempty"`

	// DownwardAPI represents the downward API about the DeviceLink.¬
	// +optional
	DownwardAPI *DeviceLinkReferenceDownwardAPISource `json:"downwardAPI,omitempty"`
}

// DeviceLinkReference defines the parameter that should be passed to the adaptor during connecting.
type DeviceLinkReference struct {
	DeviceLinkReferenceSource `json:",inline"`

	// Specifies the name of the parameter.
	Name string `json:"name,omitempty"`
}

// DeviceAdaptor defines the properties of device adaptor
type DeviceAdaptor struct {
	// Specifies the node of adaptor to be matched.
	// +optional
	Node string `json:"node,omitempty"`

	// Specifies the name of adaptor to be used.
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// [Deprecated] Specifies the parameter of adaptor to be used.
	// This field has been deprecated, it should define the connection parameter
	// as a part of device model.
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

	// Specifies the references of device to be used.
	// +optional
	References []DeviceLinkReference `json:"references,omitempty"`

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

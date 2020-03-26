package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BluetoothDeviceSpec defines the desired state of BluetoothDevice
type BluetoothDeviceSpec struct {
	Properties []DeviceProperty `json:"properties,omitempty"`
	Name       string           `json:"name,omitempty"`
	MacAddress string           `json:"macAddress,omitempty"`
}

// DeviceProperty defines an individual ble device property
type DeviceProperty struct {
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
	AccessMode  PropertyAccessMode `json:"accessMode,omitempty"`
	Visitor     PropertyVisitor    `json:"visitor,omitempty"`
}

// The access mode for  a device property.
type PropertyAccessMode string

// Access mode constants for a device property.
const (
	ReadWrite  PropertyAccessMode = "ReadWrite"
	ReadOnly   PropertyAccessMode = "ReadOnly"
	NotifyOnly PropertyAccessMode = "NotifyOnly"
)

// PropertyVisitor defines the specifics of accessing a particular device property
type PropertyVisitor struct {
	CharacteristicUUID     string                 `json:"characteristicUUID,omitempty"`
	DefaultValue           string                 `json:"defaultValue,omitempty"`
	DataWriteTo            map[string][]byte      `json:"dataWrite,omitempty"`
	BluetoothDataConverter BluetoothDataConverter `json:"dataConverter,omitempty"`
}

// BluetoothDataConverter defines the read data converting operation
type BluetoothDataConverter struct {
	StartIndex        int                   `json:"startIndex,omitempty"`
	EndIndex          int                   `json:"endIndex,omitempty"`
	ShiftLeft         int                   `json:"shiftLeft,omitempty"`
	ShiftRight        int                   `json:"shiftRight,omitempty"`
	OrderOfOperations []BluetoothOperations `json:"orderOfOperations,omitempty"`
}

type BluetoothOperations struct {
	OperationType  ArithOperationType `json:"operationType,omitempty"`
	OperationValue string             `json:"operationValue,omitempty"`
}

type ArithOperationType string

const (
	OperationAdd      ArithOperationType = "Add"
	OperationSubtract ArithOperationType = "Subtract"
	OperationMultiply ArithOperationType = "Multiply"
	OperationDivide   ArithOperationType = "Divide"
)

// BluetoothDeviceStatus defines the observed state of BluetoothDevice
type BluetoothDeviceStatus struct {
	Properties []StatusProperties `json:"properties,omitempty"`
}

type StatusProperties struct {
	Name      string      `json:"name,omitempty"`
	Desired   string      `json:"desired,omitempty"`
	Reported  string      `json:"reported,omitempty"`
	UpdatedAt metav1.Time `json:"updatedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=ble
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="MacAddress",type=integer,JSONPath=`.spec.macAddress`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// BluetoothDevice is the Schema for the ble device API
type BluetoothDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BluetoothDeviceSpec   `json:"spec,omitempty"`
	Status BluetoothDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// BluetoothDeviceList contains a list of BLE devices
type BluetoothDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BluetoothDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BluetoothDevice{}, &BluetoothDeviceList{})
}

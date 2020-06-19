package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BluetoothDeviceSpec defines the desired state of BluetoothDevice
type BluetoothDeviceSpec struct {
	// Parameter of the modbus device.
	// +optional
	Parameters *Parameters `json:"parameters,omitempty"`

	// Protocol for accessing the BLE device.
	// +kubebuilder:validation:Required
	Protocol DeviceProtocol `json:"protocol"`

	// Specifies the properties of the BLE device.
	// +optional
	Properties []DeviceProperty `json:"properties,omitempty"`

	// Specifies the extension of device.
	// +optional
	Extension *DeviceExtensionSpec `json:"extension,omitempty"`
}

type Parameters struct {
	// Specifies default device sync interval
	// +kubebuilder:default:15s
	SyncInterval v1.Duration `json:"syncInterval,omitempty"`

	// Specifies default device connection timeout
	// +kubebuilder:default:10s
	Timeout v1.Duration `json:"timeout,omitempty"`
}

// DeviceProperty defines an individual ble device property
type DeviceProperty struct {
	// The device property name.
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// The device property description.
	// +optional
	Description string `json:"description,omitempty"`
	// Required: The URL for opc-ua server endpoint.
	// +kubebuilder:validation:Required
	AccessMode PropertyAccessMode `json:"accessMode"`
	// PropertyVisitor represents the way to access the property.
	// +optional
	Visitor PropertyVisitor `json:"visitor"`
}

// DeviceProtocol defines how to connect the BLE device
type DeviceProtocol struct {
	Name       string `json:"name,omitempty"`
	MacAddress string `json:"macAddress,omitempty"`
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
	CharacteristicUUID     string                 `json:"characteristicUUID"`
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
	// Reports the extension of device.
	// +optional
	Extension *DeviceExtensionStatus `json:"extension,omitempty"`

	// Reports the status of the BLE device.
	// +optional
	Properties []StatusProperties `json:"properties,omitempty"`
}

type StatusProperties struct {
	Name       string             `json:"name,omitempty"`
	Value      string             `json:"value,omitempty"`
	AccessMode PropertyAccessMode `json:"accessMode,omitempty"`
	UpdatedAt  metav1.Time        `json:"updatedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=ble
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.protocol.name`
// +kubebuilder:printcolumn:name="MacAddress",type=integer,JSONPath=`.spec.protocol.macAddress`
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

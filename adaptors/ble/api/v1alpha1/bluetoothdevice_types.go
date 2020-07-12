package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BluetoothDeviceParameters defines the desired parameters of BluetoothDevice.
type BluetoothDeviceParameters struct {
	// Specifies default device sync interval
	// +kubebuilder:default:15s
	SyncInterval v1.Duration `json:"syncInterval,omitempty"`

	// Specifies default device connection timeout
	// +kubebuilder:default:10s
	Timeout v1.Duration `json:"timeout,omitempty"`
}

// BluetoothDeviceProtocol defines the desired protocol of BluetoothDevice.
type BluetoothDeviceProtocol struct {
	// Specifies the endpoint of device,
	// it can be the name or MAC address of device.
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`
}

// BluetoothDevicePropertyAccessMode defines the access mode of device property.
// +kubebuilder:validation:Enum=ReadWrite;ReadOnly;NotifyOnly
type BluetoothDevicePropertyAccessMode string

const (
	BluetoothDevicePropertyReadWrite  BluetoothDevicePropertyAccessMode = "ReadWrite"
	BluetoothDevicePropertyReadOnly   BluetoothDevicePropertyAccessMode = "ReadOnly"
	BluetoothDevicePropertyNotifyOnly BluetoothDevicePropertyAccessMode = "NotifyOnly"
)

// BluetoothDevicePropertyVisitor defines the specifics of accessing a particular device property
type BluetoothDevicePropertyVisitor struct {
	// Specifies the characteristic UUID of property.
	// +kubebuilder:validation:Required
	CharacteristicUUID string `json:"characteristicUUID"`

	// Specifies the default value of property,
	// when access mode is "ReadWrite".
	// +optional
	DefaultValue string `json:"defaultValue,omitempty"`

	// Specifies the data to write to device.
	// +optional
	DataWrite map[string][]byte `json:"dataWrite,omitempty"`

	// Specifies the converter to convert data read from device to a string.
	// +optional
	DataConverter BluetoothDataConverter `json:"dataConverter,omitempty"`
}

// BluetoothDeviceArithmeticOperationType defines the type of arithmetic operation.
// +kubebuilder:validation:Enum=Add;Subtract;Multiply;Divide
type BluetoothDeviceArithmeticOperationType string

const (
	BluetoothDeviceArithmeticAdd      BluetoothDeviceArithmeticOperationType = "Add"
	BluetoothDeviceArithmeticSubtract BluetoothDeviceArithmeticOperationType = "Subtract"
	BluetoothDeviceArithmeticMultiply BluetoothDeviceArithmeticOperationType = "Multiply"
	BluetoothDeviceArithmeticDivide   BluetoothDeviceArithmeticOperationType = "Divide"
)

// BluetoothDeviceArithmeticOperation defines the arithmetic operation of BluetoothDevice.
type BluetoothDeviceArithmeticOperation struct {
	// Specifies the type of arithmetic operation.
	// +kubebuilder:validation:Required
	Type BluetoothDeviceArithmeticOperationType `json:"type"`

	// Specifies the value for arithmetic operation, which is in form of float string.
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// BluetoothDataConverter defines the read data converting operation.
type BluetoothDataConverter struct {
	// Specifies the start index of the incoming byte stream to be converted.
	// +optional
	StartIndex int `json:"startIndex,omitempty"`

	// Specifies the end index of incoming byte stream to be converted.
	// +optional
	EndIndex int `json:"endIndex,omitempty"`

	// Specifies the number of bits to shift left.
	// +optional
	ShiftLeft int `json:"shiftLeft,omitempty"`

	// Specifies the number of bits to shift right.
	// +optional
	ShiftRight int `json:"shiftRight,omitempty"`

	// Specifies the operations in order if needed.
	// +listType=atomic
	// +optional
	OrderOfOperations []BluetoothDeviceArithmeticOperation `json:"orderOfOperations,omitempty"`
}

// BluetoothDeviceProperty defines an individual ble device property
type BluetoothDeviceProperty struct {
	// Specifies the name of property.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Specifies the description of property.
	// +optional
	Description string `json:"description,omitempty"`

	// Specifies the access mode of property.
	// The default value is "ReadOnly".
	// +kubebuilder:default="ReadOnly"
	AccessMode BluetoothDevicePropertyAccessMode `json:"accessMode,omitempty"`

	// Specifies the visitor of property.
	// +kubebuilder:validation:Required
	Visitor BluetoothDevicePropertyVisitor `json:"visitor"`
}

// BluetoothDeviceStatusProperty defines the observed property of BluetoothDevice.
type BluetoothDeviceStatusProperty struct {
	// Reports the properties of device.
	// +optional
	Name string `json:"name,omitempty"`

	// Reports the value of property.
	// +optional
	Value string `json:"value,omitempty"`

	// Specifies the access mode of property.
	// +optional
	AccessMode BluetoothDevicePropertyAccessMode `json:"accessMode,omitempty"`

	// Reports the updated timestamp of property.
	// +optional
	UpdatedAt *metav1.Time `json:"updatedAt,omitempty"`
}

// BluetoothDeviceSpec defines the desired state of BluetoothDevice
type BluetoothDeviceSpec struct {
	// Specifies the extension of device.
	// +optional
	Extension *BluetoothDeviceExtension `json:"extension,omitempty"`

	// Specifies the parameters of device.
	// +optional
	Parameters BluetoothDeviceParameters `json:"parameters,omitempty"`

	// Specifies the protocol for accessing the device.
	// +kubebuilder:validation:Required
	Protocol BluetoothDeviceProtocol `json:"protocol"`

	// Specifies the properties of device.
	// +listType=map
	// +listMapKey=name
	// +optional
	Properties []BluetoothDeviceProperty `json:"properties,omitempty"`
}

// BluetoothDeviceStatus defines the observed state of BluetoothDevice.
type BluetoothDeviceStatus struct {
	// Reports the status of the BLE device.
	// +optional
	Properties []BluetoothDeviceStatusProperty `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=ble
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ENDPOINT",type="string",JSONPath=`.spec.protocol.endpoint`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=`.metadata.creationTimestamp`
// BluetoothDevice is the schema for the BLE device API.
type BluetoothDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BluetoothDeviceSpec   `json:"spec,omitempty"`
	Status BluetoothDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// BluetoothDeviceList contains a list of BluetoothDevice.
type BluetoothDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []BluetoothDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BluetoothDevice{}, &BluetoothDeviceList{})
}

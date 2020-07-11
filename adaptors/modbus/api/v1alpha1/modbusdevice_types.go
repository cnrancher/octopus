package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ModbusDeviceRegisterType defines the type for the register to read a device property.
// +kubebuilder:validation:Enum=CoilRegister;DiscreteInputRegister;InputRegister;HoldingRegister
type ModbusDeviceRegisterType string

const (
	ModbusDeviceCoilRegister          ModbusDeviceRegisterType = "CoilRegister"
	ModbusDeviceDiscreteInputRegister ModbusDeviceRegisterType = "DiscreteInputRegister"
	ModbusDeviceInputRegister         ModbusDeviceRegisterType = "InputRegister"
	ModbusDeviceHoldingRegister       ModbusDeviceRegisterType = "HoldingRegister"
)

// ModbusDevicePropertyType defines the type of the property value.
// +kubebuilder:validation:Enum=int;float;string;boolean
type ModbusDevicePropertyType string

const (
	ModbusDevicePropertyTypeInt     ModbusDevicePropertyType = "int"
	ModbusDevicePropertyTypeFloat   ModbusDevicePropertyType = "float"
	ModbusDevicePropertyTypeString  ModbusDevicePropertyType = "string"
	ModbusDevicePropertyTypeBoolean ModbusDevicePropertyType = "boolean"
)

// ModbusDeviceParameters defines the desired parameters of ModbusDevice.
type ModbusDeviceParameters struct {
	// Specifies the amount of interval that synchronized to limb.
	// The default value is "15s".
	// +kubebuilder:default="15s"
	SyncInterval v1.Duration `json:"syncInterval,omitempty"`

	// Specifies the amount of timeout.
	// The default value is "10s".
	// +kubebuilder:default="10s"
	Timeout v1.Duration `json:"timeout,omitempty"`
}

func (in *ModbusDeviceParameters) GetSyncInterval() time.Duration {
	if in != nil {
		if duration := in.SyncInterval.Duration; duration > 0 {
			return duration
		}
	}
	return 15 * time.Second
}

func (in *ModbusDeviceParameters) GetTimeout() time.Duration {
	if in != nil {
		if duration := in.Timeout.Duration; duration > 0 {
			return duration
		}
	}
	return 10 * time.Second
}

// ModbusDeviceProtocol defines the desired protocol of ModbusDevice.
type ModbusDeviceProtocol struct {
	// Specifies the connection protocol as RTU
	// +optional
	RTU *ModbusDeviceProtocolRTU `json:"rtu,omitempty"`

	// Specifies the connection protocol as TCP
	// +optional
	TCP *ModbusDeviceProtocolTCP `json:"tcp,omitempty"`
}

// ModbusDeviceProtocolTCP defines the TCP protocol of ModbusDevice.
type ModbusDeviceProtocolTCP struct {
	// Specifies the IP address of device,
	// which is in form of "ip:port".
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// Specifies the worker ID of device.
	// +kubebuilder:validation:Required
	WorkerID int `json:"workerID"`
}

// ModbusDeviceProtocolRTU defines the RTU protocol of ModbusDevice.
type ModbusDeviceProtocolRTU struct {
	// Specifies the serial port of device,
	// which is in form of "/dev/ttyS0".
	// +kubebuilder:validation:Pattern="^/.*[^/]$"
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// Specifies the worker ID of device.
	// +kubebuilder:validation:Required
	WorkerID int `json:"workerID"`

	// Specifies the baud rate of connection, a measurement of transmission speed.
	// The default value is "19200".
	// +kubebuilder:default=19200
	// +optional
	BaudRate int `json:"baudRate,omitempty"`

	// Specifies the data bit of connection, selected from [5, 6, 7, 8].
	// The default value is "8".
	// +kubebuilder:validation:Enum=5;6;7;8
	// +kubebuilder:default=8
	DataBits int `json:"dataBits,omitempty"`

	// Specifies the parity of connection, selected from [N - None, E - Even, O - Odd],
	// the use of N(None) parity requires 2 stop bits.
	// The default value is "E".
	// +kubebuilder:validation:Enum=N;E;O
	// +kubebuilder:default="E"
	Parity string `json:"parity,omitempty"`

	// Specifies the stop bit of connection, selected from [1, 2],
	// the use of N(None) parity requires 2 stop bits.
	// The default value is "1".
	// +kubebuilder:validation:Enum=1;2
	// +kubebuilder:default=1
	StopBits int `json:"stopBits,omitempty"`
}

// ModbusDeviceArithmeticOperationType defines the type of arithmetic operation.
// +kubebuilder:validation:Enum=Add;Subtract;Multiply;Divide
type ModbusDeviceArithmeticOperationType string

const (
	ModbusDeviceArithmeticAdd      ModbusDeviceArithmeticOperationType = "Add"
	ModbusDeviceArithmeticSubtract ModbusDeviceArithmeticOperationType = "Subtract"
	ModbusDeviceArithmeticMultiply ModbusDeviceArithmeticOperationType = "Multiply"
	ModbusDeviceArithmeticDivide   ModbusDeviceArithmeticOperationType = "Divide"
)

// ModbusDeviceArithmeticOperation defines the arithmetic operation of ModbusDevice.
type ModbusDeviceArithmeticOperation struct {
	// Specifies the type of arithmetic operation.
	// +kubebuilder:validation:Required
	Type ModbusDeviceArithmeticOperationType `json:"type"`

	// Specifies the value for arithmetic operation, which is in form of float string.
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// ModbusDevicePropertyVisitor defines the visitor of property.
type ModbusDevicePropertyVisitor struct {
	// Specifies the register to visit.
	// +kubebuilder:validation:Required
	Register ModbusDeviceRegisterType `json:"register"`

	// Specifies the starting offset of register for read/write data.
	// +kubebuilder:validation:Required
	Offset uint16 `json:"offset"`

	// Specifies the quantity of register.
	// +kubebuilder:validation:Required
	Quantity uint16 `json:"quantity"`

	// Specifies the operations in order if needed.
	// +listType=atomic
	// +optional
	OrderOfOperations []ModbusDeviceArithmeticOperation `json:"orderOfOperations,omitempty"`
}

// ModbusDeviceProperty defines the desired property of ModbusDevice.
type ModbusDeviceProperty struct {
	// Specifies the name of property.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Specifies the description of property.
	// +optional
	Description string `json:"description,omitempty"`

	// Specifies the type of property.
	// +kubebuilder:validation:Required
	Type ModbusDevicePropertyType `json:"type"`

	// Specifies the visitor of property.
	// +kubebuilder:validation:Required
	Visitor ModbusDevicePropertyVisitor `json:"visitor"`

	// Specifies if the property is readonly.
	// The default value is "false".
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`

	// Specifies the value of property, only available in the writable property.
	// +optional
	Value string `json:"value,omitempty"`
}

// ModbusDeviceSpec defines the desired state of ModbusDevice.
type ModbusDeviceSpec struct {
	// Specifies the extension of device.
	// +optional
	Extension *ModbusDeviceExtension `json:"extension,omitempty"`

	// Specifies the parameters of device.
	// +optional
	Parameters *ModbusDeviceParameters `json:"parameters,omitempty"`

	// Specifies the protocol for accessing the device.
	// +kubebuilder:validation:Required
	Protocol ModbusDeviceProtocol `json:"protocol"`

	// Specifies the properties of device.
	// +listType=map
	// +listMapKey=name
	// +optional
	Properties []ModbusDeviceProperty `json:"properties,omitempty"`
}

// ModbusDeviceStatus defines the observed state of ModbusDevice.
type ModbusDeviceStatus struct {
	// Reports the properties of device.
	// +optional
	Properties []ModbusDeviceStatusProperty `json:"properties,omitempty"`
}

// ModbusDeviceStatusProperty defines the observed property of ModbusDevice.
type ModbusDeviceStatusProperty struct {
	// Reports the name of property.
	// +optional
	Name string `json:"name,omitempty"`

	// Reports the type of property.
	// +optional
	Type ModbusDevicePropertyType `json:"type,omitempty"`

	// Reports the value of property.
	// +optional
	Value string `json:"value,omitempty"`

	// Reports the updated timestamp of property.
	// +optional
	UpdatedAt *metav1.Time `json:"updatedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=modbus
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ENDPOINT",type="string",JSONPath=`.spec.protocol..endpoint`
// +kubebuilder:printcolumn:name="WORKER ID",type="string",JSONPath=`.spec.protocol..workerID`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=`.metadata.creationTimestamp`
// ModbusDevice is the schema for the Modbus device API.
type ModbusDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModbusDeviceSpec   `json:"spec,omitempty"`
	Status ModbusDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ModbusDeviceList contains a list of Modbus devices.
type ModbusDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ModbusDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ModbusDevice{}, &ModbusDeviceList{})
}

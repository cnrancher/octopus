package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Modbus protocol register types
	ModbusRegisterTypeCoilRegister          ModbusRegisterType = "CoilRegister"
	ModbusRegisterTypeDiscreteInputRegister ModbusRegisterType = "DiscreteInputRegister"
	ModbusRegisterTypeInputRegister         ModbusRegisterType = "InputRegister"
	ModbusRegisterTypeHoldingRegister       ModbusRegisterType = "HoldingRegister"

	//Modbus property data types
	PropertyDataTypeInt     PropertyDataType = "int"
	PropertyDataTypeString  PropertyDataType = "string"
	PropertyDataTypeFloat   PropertyDataType = "float"
	PropertyDataTypeBoolean PropertyDataType = "boolean"

	DefaultSyncInterval = 5 * time.Second
	DefaultTimeout      = 10 * time.Second
)

// The Modbus register type to read a device property.
// +kubebuilder:validation:Enum=CoilRegister;DiscreteInputRegister;InputRegister;HoldingRegister
type ModbusRegisterType string

// The property data type.
// +kubebuilder:validation:Enum=float;int;string;boolean
type PropertyDataType string

// ModbusDeviceSpec defines the desired state of ModbusDevice
type ModbusDeviceSpec struct {
	Parameters     *Parameters           `json:"parameters,omitempty"`
	ProtocolConfig *ModbusProtocolConfig `json:"protocol"`
	Properties     []DeviceProperty      `json:"properties,omitempty"`
}

type Parameters struct {
	SyncInterval v1.Duration `json:"syncInterval,omitempty"`
	Timeout      v1.Duration `json:"timeout,omitempty"`
}

// Only one of its members may be specified.
type ModbusProtocolConfig struct {
	RTU *ModbusConfigRTU `json:"rtu,omitempty"`
	TCP *ModbusConfigTCP `json:"tcp,omitempty"`
}

type ModbusConfigTCP struct {
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	SlaveID int    `json:"slaveID"`
}

type ModbusConfigRTU struct {
	// Device path (/dev/ttyS0)
	SerialPort string `json:"serialPort"`
	SlaveID    int    `json:"slaveID"`
	// Baud rate (default 19200)
	BaudRate int `json:"baudRate,omitempty"`
	// Data bits: 5, 6, 7 or 8 (default 8)
	// +kubebuilder:validation:Enum=5;6;7;8
	DataBits int `json:"dataBits,omitempty"`
	// The parity. N - None, E - Even, O - Odd, default E.
	// +kubebuilder:validation:Enum=O;E;N
	Parity string `json:"parity,omitempty"`
	// Stop bits: 1 or 2 (default 1)
	// +kubebuilder:validation:Enum=1;2
	StopBits int `json:"stopBits,omitempty"`
}

// DeviceProperty describes an individual device property / attribute like temperature / humidity etc.
type DeviceProperty struct {
	// The device property name.
	Name string `json:"name"`
	// The device property description.
	Description string `json:"description,omitempty"`
	ReadOnly    bool   `json:"readOnly,omitempty"`
	// PropertyDataType represents the type and data validation of the property.
	DataType PropertyDataType `json:"dataType"`
	Visitor  PropertyVisitor  `json:"visitor"`
	Value    string           `json:"value,omitempty"`
}

type PropertyVisitor struct {
	// Type of register
	Register ModbusRegisterType `json:"register"`
	// Offset indicates the starting register number to read/write data.
	Offset uint16 `json:"offset"`
	// The quantity of registers
	Quantity          uint16             `json:"quantity"`
	OrderOfOperations []ModbusOperations `json:"orderOfOperations,omitempty"`
}

type ModbusOperations struct {
	OperationType  ArithOperationType `json:"operationType,omitempty"`
	OperationValue string             `json:"operationValue,omitempty"`
}

// +kubebuilder:validation:Enum=Add;Subtract;Multiply;Divide
type ArithOperationType string

const (
	OperationAdd      ArithOperationType = "Add"
	OperationSubtract ArithOperationType = "Subtract"
	OperationMultiply ArithOperationType = "Multiply"
	OperationDivide   ArithOperationType = "Divide"
)

// ModbusDeviceStatus defines the observed state of ModbusDevice
type ModbusDeviceStatus struct {
	Properties []StatusProperties `json:"properties,omitempty"`
}

type StatusProperties struct {
	Name      string           `json:"name,omitempty"`
	Value     string           `json:"value,omitempty"`
	DataType  PropertyDataType `json:"dataType,omitempty"`
	UpdatedAt metav1.Time      `json:"updatedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="IP",type="string",JSONPath=".spec.protocol.tcp.ip"
// +kubebuilder:printcolumn:name="PORT",type="integer",JSONPath=".spec.protocol.tcp.port"
// +kubebuilder:printcolumn:name="SERIAL PORT",type="string",JSONPath=".spec.protocol.rtu.serialPort"
// ModbusDevice is the Schema for the modbus device API
type ModbusDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModbusDeviceSpec   `json:"spec,omitempty"`
	Status ModbusDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ModbusDeviceList contains a list of modbus devices
type ModbusDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModbusDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ModbusDevice{}, &ModbusDeviceList{})
}

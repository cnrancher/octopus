package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	endianapi "github.com/rancher/octopus/pkg/endian/api"
)

// ModbusDevicePropertyAccessMode defines the mode for accessing a device property,
// default is "ReadMany".
// +kubebuilder:validation:Enum=WriteOnce;WriteMany;ReadOnce;ReadMany
type ModbusDevicePropertyAccessMode string

const (
	ModbusDevicePropertyAccessModeWriteOnce ModbusDevicePropertyAccessMode = "WriteOnce"
	ModbusDevicePropertyAccessModeWriteMany ModbusDevicePropertyAccessMode = "WriteMany"
	ModbusDevicePropertyAccessModeReadOnce  ModbusDevicePropertyAccessMode = "ReadOnce"
	ModbusDevicePropertyAccessModeReadMany  ModbusDevicePropertyAccessMode = "ReadMany"
)

// ModbusDevicePropertyRegisterType defines the type for the register to read a device property.
// +kubebuilder:validation:Enum=CoilRegister;DiscreteInputRegister;InputRegister;HoldingRegister
type ModbusDevicePropertyRegisterType string

const (
	ModbusDevicePropertyRegisterTypeCoilRegister          ModbusDevicePropertyRegisterType = "CoilRegister"
	ModbusDevicePropertyRegisterTypeDiscreteInputRegister ModbusDevicePropertyRegisterType = "DiscreteInputRegister"
	ModbusDevicePropertyRegisterTypeInputRegister         ModbusDevicePropertyRegisterType = "InputRegister"
	ModbusDevicePropertyRegisterTypeHoldingRegister       ModbusDevicePropertyRegisterType = "HoldingRegister"
)

// ModbusDevicePropertyType defines the type of the property value.
// +kubebuilder:validation:Enum=int8;int16;int;int32;int64;uint8;uint16;uint;uint32;uint64;float;float32;double;float64;boolean;string;hexString;binaryString;base64String
type ModbusDevicePropertyType string

const (
	/*
		arithmetic types
	*/

	ModbusDevicePropertyTypeInt8    ModbusDevicePropertyType = "int8"
	ModbusDevicePropertyTypeInt16   ModbusDevicePropertyType = "int16"
	ModbusDevicePropertyTypeInt     ModbusDevicePropertyType = "int" // as same as int32
	ModbusDevicePropertyTypeInt32   ModbusDevicePropertyType = "int32"
	ModbusDevicePropertyTypeInt64   ModbusDevicePropertyType = "int64"
	ModbusDevicePropertyTypeUint8   ModbusDevicePropertyType = "uint8"
	ModbusDevicePropertyTypeUint16  ModbusDevicePropertyType = "uint16"
	ModbusDevicePropertyTypeUint    ModbusDevicePropertyType = "uint" // as same as uint32
	ModbusDevicePropertyTypeUint32  ModbusDevicePropertyType = "uint32"
	ModbusDevicePropertyTypeUint64  ModbusDevicePropertyType = "uint64"
	ModbusDevicePropertyTypeFloat   ModbusDevicePropertyType = "float" // as same as float32
	ModbusDevicePropertyTypeFloat32 ModbusDevicePropertyType = "float32"
	ModbusDevicePropertyTypeDouble  ModbusDevicePropertyType = "double" // as same as float64
	ModbusDevicePropertyTypeFloat64 ModbusDevicePropertyType = "float64"

	/*
		none arithmetic types
	*/

	ModbusDevicePropertyTypeBoolean ModbusDevicePropertyType = "boolean"
	ModbusDevicePropertyTypeString  ModbusDevicePropertyType = "string"

	/*
		for bytes
	*/
	ModbusDevicePropertyTypeHexString    ModbusDevicePropertyType = "hexString"
	ModbusDevicePropertyTypeBinaryString ModbusDevicePropertyType = "binaryString"
	ModbusDevicePropertyTypeBase64String ModbusDevicePropertyType = "base64String"
)

// ModbusDeviceProtocol defines the desired protocol of ModbusDevice.
type ModbusDeviceProtocol struct {
	// Specifies the connection protocol as RTU.
	// +optional
	RTU *ModbusDeviceProtocolRTU `json:"rtu,omitempty"`

	// Specifies the connection protocol as TCP.
	// +optional
	TCP *ModbusDeviceProtocolTCP `json:"tcp,omitempty"`
}

func (in *ModbusDeviceProtocol) GetSyncInterval() time.Duration {
	if in != nil {
		if in.TCP != nil {
			return in.TCP.GetSyncInterval()
		}
		if in.RTU != nil {
			return in.RTU.GetSyncInterval()
		}
	}
	return 10 * time.Second
}

type ModbusDeviceProtocolParameters struct {
	// Specifies the amount of interval for synchronizing the device.
	// The default value is "10s".
	// +kubebuilder:default="10s"
	SyncInterval metav1.Duration `json:"syncInterval,omitempty"`

	// Specifies the amount of timeout for connecting to the device.
	// The default value is "10s".
	// +kubebuilder:default="10s"
	ConnectTimeout metav1.Duration `json:"connectTimeout,omitempty"`
}

func (in *ModbusDeviceProtocolParameters) GetSyncInterval() time.Duration {
	if in != nil {
		if duration := in.SyncInterval.Duration; duration > 0 {
			return duration
		}
	}
	return 10 * time.Second
}

func (in *ModbusDeviceProtocolParameters) GetConnectTimeout() time.Duration {
	if in != nil {
		if duration := in.ConnectTimeout.Duration; duration > 0 {
			return duration
		}
	}
	return 10 * time.Second
}

// ModbusDeviceProtocolTCP defines the TCP protocol of ModbusDevice.
type ModbusDeviceProtocolTCP struct {
	ModbusDeviceProtocolParameters `json:",inline"`

	// Specifies the IP address of device,
	// which is in form of "ip:port".
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// Specifies the worker ID of device,
	// it's from 1 to 247.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=247
	// +kubebuilder:validation:Required
	WorkerID int `json:"workerID"`
}

// ModbusDeviceProtocolRTU defines the RTU protocol of ModbusDevice.
type ModbusDeviceProtocolRTU struct {
	ModbusDeviceProtocolParameters `json:",inline"`

	// Specifies the serial port of device,
	// which is in form of "/dev/ttyS0".
	// +kubebuilder:validation:Pattern="^/.*[^/]$"
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// Specifies the worker ID of device,
	// it's from 1 to 247.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=247
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

// ModbusDevicePropertyValueArithmeticOperationType defines the type of arithmetic operation.
// +kubebuilder:validation:Enum=Add;Subtract;Multiply;Divide
type ModbusDevicePropertyValueArithmeticOperationType string

const (
	ModbusDevicePropertyValueArithmeticAdd      ModbusDevicePropertyValueArithmeticOperationType = "Add"
	ModbusDevicePropertyValueArithmeticSubtract ModbusDevicePropertyValueArithmeticOperationType = "Subtract"
	ModbusDevicePropertyValueArithmeticMultiply ModbusDevicePropertyValueArithmeticOperationType = "Multiply"
	ModbusDevicePropertyValueArithmeticDivide   ModbusDevicePropertyValueArithmeticOperationType = "Divide"
)

// ModbusDevicePropertyValueArithmeticOperation defines the arithmetic operation of ModbusDevice.
type ModbusDevicePropertyValueArithmeticOperation struct {
	// Specifies the type of arithmetic operation.
	// +kubebuilder:validation:Required
	Type ModbusDevicePropertyValueArithmeticOperationType `json:"type"`

	// Specifies the value for arithmetic operation, which is in form of float string.
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// ModbusDevicePropertyVisitor defines the visitor of property.
type ModbusDevicePropertyVisitor struct {
	// Specifies the register to visit.
	// +kubebuilder:validation:Required
	Register ModbusDevicePropertyRegisterType `json:"register"`

	// Specifies the starting offset of register for read/write data.
	// +kubebuilder:validation:Required
	Offset uint16 `json:"offset"`

	// Specifies the quantity of register,
	// the corresponding type restrictions are as follows:
	// - when the register is CoilRegister, quantity is not longer than 1968;
	// - when the register is HoldingRegister, quantity is not longer than 123.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Quantity uint16 `json:"quantity,omitempty"`

	// Specifies the endianness of value, only available in basic types.
	// +kubebuilder:default="BigEndian"
	Endianness endianapi.DevicePropertyValueEndianness `json:"endianness,omitempty"`

	// Specifies the arithmetic operations in order if needed, only available in arithmetic types.
	// +listType=atomic
	// +optional
	ArithmeticOperations []ModbusDevicePropertyValueArithmeticOperation `json:"arithmeticOperations,omitempty"`

	// Specifies the precision of the arithmetic operation result.
	// The default is "2".
	// +optional
	ArithmeticOperationPrecision *int `json:"arithmeticOperationPrecision,omitempty"`
}

func (in *ModbusDevicePropertyVisitor) GetArithmeticOperationPrecision() int {
	if in != nil && in.ArithmeticOperationPrecision != nil {
		return *in.ArithmeticOperationPrecision
	}
	return 2
}

func (in *ModbusDevicePropertyVisitor) GetEndianness() endianapi.DevicePropertyValueEndianness {
	if in != nil && string(in.Endianness) != "" {
		return in.Endianness
	}
	return endianapi.DevicePropertyValueEndiannessBigEndian
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

	// Specifies the access mode of property.
	// The default value is "ReadMany".
	// +listType=set
	// +kubebuilder:default={ReadMany}
	AccessModes []ModbusDevicePropertyAccessMode `json:"accessModes,omitempty"`

	// Specifies the visitor of property.
	// +kubebuilder:validation:Required
	Visitor ModbusDevicePropertyVisitor `json:"visitor"`

	// Specifies the value of property, only available in the writable property.
	// +optional
	Value string `json:"value,omitempty"`
}

// MergeAccessModes merges the duplicated modes and then returns the access mode array.
func (in *ModbusDeviceProperty) MergeAccessModes() []ModbusDevicePropertyAccessMode {
	if in != nil && len(in.AccessModes) != 0 {
		// NB(thxCode) if both "*Once" and "*Many" are specified,
		// we can merge "*Once" to "*Many" via bitmap,
		// and keep "Write*" before "Read*".
		var mode byte
		for _, accessMode := range in.AccessModes {
			switch accessMode {
			case ModbusDevicePropertyAccessModeWriteOnce:
				mode = mode | 0x04
			case ModbusDevicePropertyAccessModeWriteMany:
				mode = mode | 0x08
			case ModbusDevicePropertyAccessModeReadOnce:
				mode = mode | 0x01
			default: // ModbusDevicePropertyAccessModeReadMany
				mode = mode | 0x02
			}
		}
		if mode&0x08 == 0x08 {
			mode = mode & 0xfb
		}
		if mode&0x02 == 0x02 {
			mode = mode & 0xfe
		}

		switch mode {
		case 0x0a: // 1010
			return []ModbusDevicePropertyAccessMode{ModbusDevicePropertyAccessModeWriteMany, ModbusDevicePropertyAccessModeReadMany}
		case 0x09: // 1001
			return []ModbusDevicePropertyAccessMode{ModbusDevicePropertyAccessModeWriteMany, ModbusDevicePropertyAccessModeReadOnce}
		case 0x06: // 0110
			return []ModbusDevicePropertyAccessMode{ModbusDevicePropertyAccessModeWriteOnce, ModbusDevicePropertyAccessModeReadMany}
		case 0x05: // 0101
			return []ModbusDevicePropertyAccessMode{ModbusDevicePropertyAccessModeWriteOnce, ModbusDevicePropertyAccessModeReadOnce}
		case 0x04: // 0100
			return []ModbusDevicePropertyAccessMode{ModbusDevicePropertyAccessModeWriteOnce}
		case 0x08: // 1000
			return []ModbusDevicePropertyAccessMode{ModbusDevicePropertyAccessModeWriteMany}
		case 0x01: // 0001
			return []ModbusDevicePropertyAccessMode{ModbusDevicePropertyAccessModeReadOnce}
		default: // 0010
		}
	}
	return []ModbusDevicePropertyAccessMode{ModbusDevicePropertyAccessModeReadMany}
}

// ModbusDeviceSpec defines the desired state of ModbusDevice.
type ModbusDeviceSpec struct {
	// Specifies the extension of device.
	// +optional
	Extension *ModbusDeviceExtension `json:"extension,omitempty"`

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

	// Reports the access mode of property.
	// +optional
	AccessModes []ModbusDevicePropertyAccessMode `json:"accessModes,omitempty"`

	// Reports the value of property.
	// +optional
	Value string `json:"value,omitempty"`

	// Reports the operation result of property if configured `arithmeticOperations`.
	// +optional
	OperationResult string `json:"operationResult,omitempty"`

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

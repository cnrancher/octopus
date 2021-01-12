package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	endianapi "github.com/rancher/octopus/pkg/endian/api"
)

// BluetoothDevicePropertyAccessMode defines the mode for accessing a device property,
// default is "ReadMany".
// +kubebuilder:validation:Enum=Notify;WriteOnce;WriteMany;ReadOnce;ReadMany
type BluetoothDevicePropertyAccessMode string

const (
	BluetoothDevicePropertyAccessModeNotify    BluetoothDevicePropertyAccessMode = "Notify"
	BluetoothDevicePropertyAccessModeWriteOnce BluetoothDevicePropertyAccessMode = "WriteOnce"
	BluetoothDevicePropertyAccessModeWriteMany BluetoothDevicePropertyAccessMode = "WriteMany"
	BluetoothDevicePropertyAccessModeReadOnce  BluetoothDevicePropertyAccessMode = "ReadOnce"
	BluetoothDevicePropertyAccessModeReadMany  BluetoothDevicePropertyAccessMode = "ReadMany"
)

// BluetoothDevicePropertyType defines the type of the property value.
// +kubebuilder:validation:Enum=int8;int16;int;int32;int64;uint8;uint16;uint;uint32;uint64;float;float32;double;float64;boolean;string;hexString;binaryString;base64String
type BluetoothDevicePropertyType string

const (
	/*
		arithmetic types
	*/

	BluetoothDevicePropertyTypeInt8    BluetoothDevicePropertyType = "int8"
	BluetoothDevicePropertyTypeInt16   BluetoothDevicePropertyType = "int16"
	BluetoothDevicePropertyTypeInt     BluetoothDevicePropertyType = "int" // as same as int32
	BluetoothDevicePropertyTypeInt32   BluetoothDevicePropertyType = "int32"
	BluetoothDevicePropertyTypeInt64   BluetoothDevicePropertyType = "int64"
	BluetoothDevicePropertyTypeUint8   BluetoothDevicePropertyType = "uint8"
	BluetoothDevicePropertyTypeUint16  BluetoothDevicePropertyType = "uint16"
	BluetoothDevicePropertyTypeUint    BluetoothDevicePropertyType = "uint" // as same as uint32
	BluetoothDevicePropertyTypeUint32  BluetoothDevicePropertyType = "uint32"
	BluetoothDevicePropertyTypeUint64  BluetoothDevicePropertyType = "uint64"
	BluetoothDevicePropertyTypeFloat   BluetoothDevicePropertyType = "float" // as same as float32
	BluetoothDevicePropertyTypeFloat32 BluetoothDevicePropertyType = "float32"
	BluetoothDevicePropertyTypeDouble  BluetoothDevicePropertyType = "double" // as same as float64
	BluetoothDevicePropertyTypeFloat64 BluetoothDevicePropertyType = "float64"

	/*
		none arithmetic types
	*/

	BluetoothDevicePropertyTypeBoolean BluetoothDevicePropertyType = "boolean"
	BluetoothDevicePropertyTypeString  BluetoothDevicePropertyType = "string"

	/*
		for bytes
	*/
	BluetoothDevicePropertyTypeHexString    BluetoothDevicePropertyType = "hexString"
	BluetoothDevicePropertyTypeBinaryString BluetoothDevicePropertyType = "binaryString"
	BluetoothDevicePropertyTypeBase64String BluetoothDevicePropertyType = "base64String"
)

// BluetoothDeviceProtocol defines the desired protocol of BluetoothDevice.
type BluetoothDeviceProtocol struct {
	// Specifies the endpoint of GATT peripheral,
	// it can be the name or MAC address of device.
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// Specifies the amount of timeout for scanning the GATT peripheral.
	// The default is "15s".
	// +kubebuilder:default="15s"
	ScanTimeout metav1.Duration `json:"scanTimeout,omitempty"`

	// Specifies the amount of timeout for connecting to the GATT peripheral.
	// The default is "10s".
	// +kubebuilder:default="10s"
	ConnectTimeout metav1.Duration `json:"connectTimeout,omitempty"`

	// Specifies the MTU of the connection,
	// central will try to use this value to negotiate with the GATT peripheral.
	// +kubebuilder:validation:Minimum=23
	// +kubebuilder:validation:Maximum=515
	ConnectionMTU *int `json:"connectionMTU,omitempty"`

	// Specifies the amount of interval for synchronizing the GATT peripheral.
	// The default value is "10s".
	// +kubebuilder:default="10s"
	SyncInterval metav1.Duration `json:"syncInterval,omitempty"`

	// Specifies to reconnect the GATT peripheral automatically.
	// The default value is "true".
	// +kubebuilder:default=true
	AutoReconnect *bool `json:"autoReconnect,omitempty"`

	// Specifies the amount of time that the client should wait
	// before reconnecting to the GATT peripheral. The first reconnect interval is 1 second,
	// and then the interval is incremented by *2 until `MaxReconnectInterval` is reached.
	// This is only valid if `AutoReconnect` is true.
	// A duration of 0 may trigger the reconnection immediately.
	// The default value is "10m".
	// +kubebuilder:default="10m"
	MaxReconnectInterval metav1.Duration `json:"maxReconnectInterval,omitempty"`

	// Specifies to only subscribe the "notify" notifications,
	// if the value of a GATT characteristic supports to be subscribed as both notification and indication.
	// The default value is "false".
	// +optional
	OnlySubscribeNotificationValue bool `json:"onlySubscribeNotificationValue,omitempty"`

	// Specifies to only write without response,
	// if a GATT characteristic can be written with a response or no response.
	// The default value is "false".
	// +optional
	OnlyWriteValueWithoutResponse bool `json:"onlyWriteValueWithoutResponse,omitempty"`
}

func (in *BluetoothDeviceProtocol) GetScanTimeout() time.Duration {
	if in != nil {
		if duration := in.ScanTimeout.Duration; duration > 0 {
			return duration
		}
	}
	return 15 * time.Second
}

func (in *BluetoothDeviceProtocol) GetConnectTimeout() time.Duration {
	if in != nil {
		if duration := in.ConnectTimeout.Duration; duration > 0 {
			return duration
		}
	}
	return 10 * time.Second
}

func (in *BluetoothDeviceProtocol) GetConnectionMTU() int {
	if in != nil && in.ConnectionMTU != nil {
		if mtu := *in.ConnectionMTU; mtu > 0 {
			return mtu
		}
	}
	return 0
}

func (in *BluetoothDeviceProtocol) GetSyncInterval() time.Duration {
	if in != nil {
		if duration := in.SyncInterval.Duration; duration > 0 {
			return duration
		}
	}
	return 10 * time.Second
}

func (in *BluetoothDeviceProtocol) IsAutoReconnect() bool {
	if in != nil && in.AutoReconnect != nil {
		return *in.AutoReconnect
	}
	return true
}

func (in *BluetoothDeviceProtocol) GetMaxReconnectInterval() time.Duration {
	if in != nil {
		if duration := in.MaxReconnectInterval.Duration; duration > 0 {
			return duration
		}
	}
	return 10 * time.Minute
}

// BluetoothDevicePropertyValueArithmeticOperationType defines the type of arithmetic operation.
// +kubebuilder:validation:Enum=Add;Subtract;Multiply;Divide
type BluetoothDevicePropertyValueArithmeticOperationType string

const (
	BluetoothDevicePropertyValueArithmeticAdd      BluetoothDevicePropertyValueArithmeticOperationType = "Add"
	BluetoothDevicePropertyValueArithmeticSubtract BluetoothDevicePropertyValueArithmeticOperationType = "Subtract"
	BluetoothDevicePropertyValueArithmeticMultiply BluetoothDevicePropertyValueArithmeticOperationType = "Multiply"
	BluetoothDevicePropertyValueArithmeticDivide   BluetoothDevicePropertyValueArithmeticOperationType = "Divide"
)

// BluetoothDevicePropertyValueArithmeticOperation defines the arithmetic operation of property value.
type BluetoothDevicePropertyValueArithmeticOperation struct {
	// Specifies the type of arithmetic operation.
	// +kubebuilder:validation:Required
	Type BluetoothDevicePropertyValueArithmeticOperationType `json:"type"`

	// Specifies the value for arithmetic operation, which is in form of float string.
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// BluetoothDevicePropertyValueContentType defines the content type of property value,
// default is "bytes".
// +kubebuilder:validation:Enum=text;bytes
type BluetoothDevicePropertyValueContentType string

const (
	BluetoothDevicePropertyValueContentTypeText  BluetoothDevicePropertyValueContentType = "text"
	BluetoothDevicePropertyValueContentTypeBytes BluetoothDevicePropertyValueContentType = "bytes"
)

// BluetoothDevicePropertyVisitor defines the specifics of accessing a particular device property
type BluetoothDevicePropertyVisitor struct {
	// Specifies the GATT service UUID of property.
	// +optional
	Service string `json:"service,omitempty"`

	// Specifies the GATT characteristic UUID of property.
	// +kubebuilder:validation:Required
	Characteristic string `json:"characteristic"`

	// Specifies the content type of property value.
	// The default is "bytes".
	// +kubebuilder:default="bytes"
	ContentType BluetoothDevicePropertyValueContentType `json:"contentType,omitempty"`

	// Specifies the endianness of value, only available in basic bytes(content-type) types.
	// +kubebuilder:default="LittleEndian"
	Endianness endianapi.DevicePropertyValueEndianness `json:"endianness,omitempty"`

	// Specifies the arithmetic operations in order if needed, only available in arithmetic types.
	// +listType=atomic
	// +optional
	ArithmeticOperations []BluetoothDevicePropertyValueArithmeticOperation `json:"arithmeticOperations,omitempty"`

	// Specifies the precision of the arithmetic operation result.
	// +optional
	ArithmeticOperationPrecision *int `json:"arithmeticOperationPrecision,omitempty"`
}

func (in *BluetoothDevicePropertyVisitor) GetArithmeticOperationPrecision() int {
	if in != nil && in.ArithmeticOperationPrecision != nil {
		return *in.ArithmeticOperationPrecision
	}
	return 2
}

func (in *BluetoothDevicePropertyVisitor) GetEndianness() endianapi.DevicePropertyValueEndianness {
	if in != nil && string(in.Endianness) != "" {
		return in.Endianness
	}
	return endianapi.DevicePropertyValueEndiannessLittleEndian
}

// BluetoothDeviceProperty defines an individual ble device property
type BluetoothDeviceProperty struct {
	// Specifies the name of property.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Specifies the description of property.
	// +optional
	Description string `json:"description,omitempty"`

	// Specifies the type of property.
	// +kubebuilder:validation:Required
	Type BluetoothDevicePropertyType `json:"type"`

	// Specifies the access mode of property.
	// The default value is "ReadMany".
	// +listType=set
	// +kubebuilder:default={ReadMany}
	AccessModes []BluetoothDevicePropertyAccessMode `json:"accessModes,omitempty"`

	// Specifies the visitor of property.
	// +kubebuilder:validation:Required
	Visitor BluetoothDevicePropertyVisitor `json:"visitor"`

	// Specifies the value of property, only available in the writable property.
	// +optional
	Value string `json:"value,omitempty"`
}

// MergeAccessModes merges the duplicated modes and then returns the access mode array.
func (in *BluetoothDeviceProperty) MergeAccessModes() []BluetoothDevicePropertyAccessMode {
	if in != nil && len(in.AccessModes) != 0 {
		// NB(thxCode) if both "*Once" and "*Many" are specified,
		// we can merge "*Once" to "*Many" via bitmap,
		// and keep "Write*" before "Read*".
		var mode byte
		for _, accessMode := range in.AccessModes {
			switch accessMode {
			case BluetoothDevicePropertyAccessModeNotify:
				mode = mode | 0x10
			case BluetoothDevicePropertyAccessModeWriteOnce:
				mode = mode | 0x04
			case BluetoothDevicePropertyAccessModeWriteMany:
				mode = mode | 0x08
			case BluetoothDevicePropertyAccessModeReadOnce:
				mode = mode | 0x01
			default: // BluetoothDevicePropertyAccessModeReadMany
				mode = mode | 0x02
			}
		}
		if mode&0x08 == 0x08 {
			mode = mode & 0xfb
		}
		if mode&0x02 == 0x02 {
			mode = mode & 0xfe
		}

		var accessModes []BluetoothDevicePropertyAccessMode
		if mode&0x10 == 0x10 {
			accessModes = append(accessModes, BluetoothDevicePropertyAccessModeNotify)
			mode = mode & 0x0f
		}
		switch mode {
		case 0x0a: // 1010
			accessModes = append(accessModes, BluetoothDevicePropertyAccessModeWriteMany, BluetoothDevicePropertyAccessModeReadMany)
		case 0x09: // 1001
			accessModes = append(accessModes, BluetoothDevicePropertyAccessModeWriteMany, BluetoothDevicePropertyAccessModeReadOnce)
		case 0x06: // 0110
			accessModes = append(accessModes, BluetoothDevicePropertyAccessModeWriteOnce, BluetoothDevicePropertyAccessModeReadMany)
		case 0x05: // 0101
			accessModes = append(accessModes, BluetoothDevicePropertyAccessModeWriteOnce, BluetoothDevicePropertyAccessModeReadOnce)
		case 0x04: // 0100
			accessModes = append(accessModes, BluetoothDevicePropertyAccessModeWriteOnce)
		case 0x08: // 1000
			accessModes = append(accessModes, BluetoothDevicePropertyAccessModeWriteMany)
		case 0x01: // 0001
			accessModes = append(accessModes, BluetoothDevicePropertyAccessModeReadOnce)
		case 0x02: // 0010
			accessModes = append(accessModes, BluetoothDevicePropertyAccessModeReadMany)
		}

		// never reach
		return accessModes
	}
	return []BluetoothDevicePropertyAccessMode{BluetoothDevicePropertyAccessModeReadMany}
}

// BluetoothDeviceStatusProperty defines the observed property of BluetoothDevice.
type BluetoothDeviceStatusProperty struct {
	// Reports the name of property.
	// +optional
	Name string `json:"name,omitempty"`

	// Reports the type of property.
	// +optional
	Type BluetoothDevicePropertyType `json:"type,omitempty"`

	// Reports the access mode of property.
	// +optional
	AccessModes []BluetoothDevicePropertyAccessMode `json:"accessModes,omitempty"`

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

// BluetoothDeviceSpec defines the desired state of BluetoothDevice
type BluetoothDeviceSpec struct {
	// Specifies the extension of device.
	// +optional
	Extension *BluetoothDeviceExtension `json:"extension,omitempty"`

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

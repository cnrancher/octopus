package physical

import (
	"encoding/hex"
	"strconv"

	"github.com/pkg/errors"

	"github.com/rancher/octopus/adaptors/ble/api/v1alpha1"
	endianapi "github.com/rancher/octopus/pkg/endian/api"
	"github.com/rancher/octopus/pkg/util/converter"
)

// doArithmeticOperations helps to calculate the raw value with operations.
func doArithmeticOperations(raw interface{}, visitor *v1alpha1.BluetoothDevicePropertyVisitor) (string, error) {
	if len(visitor.ArithmeticOperations) == 0 {
		return "", nil
	}

	var result float64
	switch v := raw.(type) {
	case float64:
		result = v
	case string:
		var err error
		result, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse raw string to float64 value")
		}
	default:
		return "", errors.Errorf("cannot parse %T type value to float64 value", raw)
	}

	for _, executeOperation := range visitor.ArithmeticOperations {
		var operationValue, err = strconv.ParseFloat(executeOperation.Value, 64)
		if err != nil {
			return "", errors.Wrapf(err, "failed to execute arithmetic operations as error parsing %s operation's value", executeOperation.Type)
		}
		switch executeOperation.Type {
		case v1alpha1.BluetoothDevicePropertyValueArithmeticAdd:
			result = result + operationValue
		case v1alpha1.BluetoothDevicePropertyValueArithmeticSubtract:
			result = result - operationValue
		case v1alpha1.BluetoothDevicePropertyValueArithmeticMultiply:
			result = result * operationValue
		case v1alpha1.BluetoothDevicePropertyValueArithmeticDivide:
			result = result / operationValue
		}
	}
	return strconv.FormatFloat(result, 'f', visitor.GetArithmeticOperationPrecision(), 64), nil
}

// convertValueToBytes converts the property value to a byte array according to the property type.
func convertValueToBytes(prop *v1alpha1.BluetoothDeviceProperty) (b []byte, err error) {
	var visitor = &prop.Visitor

	if visitor.ContentType == v1alpha1.BluetoothDevicePropertyValueContentTypeText {
		return []byte(prop.Value), nil
	}

	switch prop.Type {
	case v1alpha1.BluetoothDevicePropertyTypeInt8:
		b, err = visitor.GetEndianness().ConvertInt8String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.BluetoothDevicePropertyTypeInt16:
		b, err = visitor.GetEndianness().ConvertInt16String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.BluetoothDevicePropertyTypeInt32, v1alpha1.BluetoothDevicePropertyTypeInt:
		b, err = visitor.GetEndianness().ConvertInt32String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.BluetoothDevicePropertyTypeInt64:
		b, err = visitor.GetEndianness().ConvertInt64String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.BluetoothDevicePropertyTypeUint8:
		b, err = visitor.GetEndianness().ConvertUint8String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.BluetoothDevicePropertyTypeUint16:
		b, err = visitor.GetEndianness().ConvertUint16String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.BluetoothDevicePropertyTypeUint32, v1alpha1.BluetoothDevicePropertyTypeUint:
		b, err = visitor.GetEndianness().ConvertUint32String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.BluetoothDevicePropertyTypeUint64:
		b, err = visitor.GetEndianness().ConvertUint64String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.BluetoothDevicePropertyTypeFloat32, v1alpha1.BluetoothDevicePropertyTypeFloat:
		b, err = visitor.GetEndianness().ConvertFloat32String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.BluetoothDevicePropertyTypeFloat64, v1alpha1.BluetoothDevicePropertyTypeDouble:
		b, err = visitor.GetEndianness().ConvertFloat64String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.BluetoothDevicePropertyTypeBoolean:
		b, err = visitor.GetEndianness().ConvertBooleanString(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.BluetoothDevicePropertyTypeHexString:
		b, err = hex.DecodeString(prop.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the hex value to bytes")
		}
	case v1alpha1.BluetoothDevicePropertyTypeBinaryString:
		b, err = converter.DecodeBinaryString(prop.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the binary value to bytes")
		}
	case v1alpha1.BluetoothDevicePropertyTypeBase64String:
		b, err = converter.DecodeBase64String(prop.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the base64 value to bytes")
		}
	case v1alpha1.BluetoothDevicePropertyTypeString:
		b, err = visitor.GetEndianness().ConvertString(prop.Value)
		if err != nil {
			return
		}
	default:
		return nil, errors.Errorf("cannot convert the %s value to bytes", prop.Type)
	}

	return
}

// parseValueFromBytes parses the property value from a byte array according to the property type,
// and returns the operation result too.
func parseValueFromBytes(data []byte, prop *v1alpha1.BluetoothDeviceProperty) (value string, result string, err error) {
	var visitor = &prop.Visitor

	if visitor.ContentType == v1alpha1.BluetoothDevicePropertyValueContentTypeText {
		if data[0] == '"' && data[len(data)-1] == '"' {
			value = converter.UnsafeBytesToString(data[1 : len(data)-1])
		} else {
			value = converter.UnsafeBytesToString(data)
		}

		// executes arithmetic operations
		switch prop.Type {
		case v1alpha1.BluetoothDevicePropertyTypeInt8,
			v1alpha1.BluetoothDevicePropertyTypeInt16,
			v1alpha1.BluetoothDevicePropertyTypeInt32, v1alpha1.BluetoothDevicePropertyTypeInt,
			v1alpha1.BluetoothDevicePropertyTypeInt64,
			v1alpha1.BluetoothDevicePropertyTypeUint8,
			v1alpha1.BluetoothDevicePropertyTypeUint16,
			v1alpha1.BluetoothDevicePropertyTypeUint32, v1alpha1.BluetoothDevicePropertyTypeUint,
			v1alpha1.BluetoothDevicePropertyTypeUint64,
			v1alpha1.BluetoothDevicePropertyTypeFloat32, v1alpha1.BluetoothDevicePropertyTypeFloat,
			v1alpha1.BluetoothDevicePropertyTypeFloat64, v1alpha1.BluetoothDevicePropertyTypeDouble:
			result, err = doArithmeticOperations(value, visitor)
		}

		return
	}

	switch prop.Type {
	case v1alpha1.BluetoothDevicePropertyTypeInt8:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseInt8StringValue(data)
		if err != nil {
			return
		}
		result, err = doArithmeticOperations(sv.Float64(), visitor)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.BluetoothDevicePropertyTypeInt16:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseInt16StringValue(data)
		if err != nil {
			return
		}
		result, err = doArithmeticOperations(sv.Float64(), visitor)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.BluetoothDevicePropertyTypeInt32, v1alpha1.BluetoothDevicePropertyTypeInt:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseInt32StringValue(data)
		if err != nil {
			return
		}
		result, err = doArithmeticOperations(sv.Float64(), visitor)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.BluetoothDevicePropertyTypeInt64:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseInt64StringValue(data)
		if err != nil {
			return
		}
		result, err = doArithmeticOperations(sv.Float64(), visitor)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.BluetoothDevicePropertyTypeUint8:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseUint8StringValue(data)
		if err != nil {
			return
		}
		result, err = doArithmeticOperations(sv.Float64(), visitor)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.BluetoothDevicePropertyTypeUint16:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseUint16StringValue(data)
		if err != nil {
			return
		}
		result, err = doArithmeticOperations(sv.Float64(), visitor)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.BluetoothDevicePropertyTypeUint32, v1alpha1.BluetoothDevicePropertyTypeUint:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseUint32StringValue(data)
		if err != nil {
			return
		}
		result, err = doArithmeticOperations(sv.Float64(), visitor)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.BluetoothDevicePropertyTypeUint64:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseUint64StringValue(data)
		if err != nil {
			return
		}
		result, err = doArithmeticOperations(sv.Float64(), visitor)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.BluetoothDevicePropertyTypeFloat32, v1alpha1.BluetoothDevicePropertyTypeFloat:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseFloat32StringValue(data)
		if err != nil {
			return
		}
		result, err = doArithmeticOperations(sv.Float64(), visitor)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.BluetoothDevicePropertyTypeFloat64, v1alpha1.BluetoothDevicePropertyTypeDouble:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseFloat64StringValue(data)
		if err != nil {
			return
		}
		result, err = doArithmeticOperations(sv.Float64(), visitor)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.BluetoothDevicePropertyTypeBoolean:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseBooleanStringValue(data)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.BluetoothDevicePropertyTypeHexString:
		value = hex.EncodeToString(data)
	case v1alpha1.BluetoothDevicePropertyTypeBinaryString:
		value = converter.EncodeBinaryToString(data)
	case v1alpha1.BluetoothDevicePropertyTypeBase64String:
		value = converter.EncodeBase64ToString(data)
	case v1alpha1.BluetoothDevicePropertyTypeString:
		value, err = visitor.GetEndianness().ParseString(data)
		if err != nil {
			return
		}
	default:
		err = errors.Errorf("cannot parse bytes to %s type", prop.Type)
	}

	return
}

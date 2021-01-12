package physical

import (
	"encoding/hex"
	"strconv"

	"github.com/pkg/errors"

	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
	endianapi "github.com/rancher/octopus/pkg/endian/api"
	"github.com/rancher/octopus/pkg/util/converter"
)

// doArithmeticOperations helps to calculate the raw value with operations.
func doArithmeticOperations(raw float64, visitor *v1alpha1.ModbusDevicePropertyVisitor) (string, error) {
	if len(visitor.ArithmeticOperations) == 0 {
		return "", nil
	}

	var result = raw
	for _, executeOperation := range visitor.ArithmeticOperations {
		operationValue, err := strconv.ParseFloat(executeOperation.Value, 64)
		if err != nil {
			return "", errors.Wrapf(err, "failed to execute arithmetic operations as error parsing %s operation's value", executeOperation.Type)
		}
		switch executeOperation.Type {
		case v1alpha1.ModbusDevicePropertyValueArithmeticAdd:
			result = result + operationValue
		case v1alpha1.ModbusDevicePropertyValueArithmeticSubtract:
			result = result - operationValue
		case v1alpha1.ModbusDevicePropertyValueArithmeticMultiply:
			result = result * operationValue
		case v1alpha1.ModbusDevicePropertyValueArithmeticDivide:
			result = result / operationValue
		}
	}
	return strconv.FormatFloat(result, 'f', visitor.GetArithmeticOperationPrecision(), 64), nil
}

// convertValueToBytes converts the property value to a byte array according to the property type.
func convertValueToBytes(prop *v1alpha1.ModbusDeviceProperty) (b []byte, err error) {
	var visitor = &prop.Visitor

	switch prop.Type {
	case v1alpha1.ModbusDevicePropertyTypeInt8:
		var val int64
		val, err = strconv.ParseInt(prop.Value, 10, 8)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to int8")
		}
		// NB(thxCode) in modbus, we use 16bits to store 8bits numeric data.
		b, err = visitor.GetEndianness().ConvertInt16(int16(val))
		if err != nil {
			return
		}
	case v1alpha1.ModbusDevicePropertyTypeInt16:
		b, err = visitor.GetEndianness().ConvertInt16String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.ModbusDevicePropertyTypeInt32, v1alpha1.ModbusDevicePropertyTypeInt:
		b, err = visitor.GetEndianness().ConvertInt32String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.ModbusDevicePropertyTypeInt64:
		b, err = visitor.GetEndianness().ConvertInt64String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.ModbusDevicePropertyTypeUint8:
		var val uint64
		val, err = strconv.ParseUint(prop.Value, 10, 8)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to uint8")
		}
		// NB(thxCode) in modbus, we use 16bits to store 8bits numeric data.
		b, err = visitor.GetEndianness().ConvertUint16(uint16(val))
		if err != nil {
			return
		}
	case v1alpha1.ModbusDevicePropertyTypeUint16:
		b, err = visitor.GetEndianness().ConvertUint16String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.ModbusDevicePropertyTypeUint32, v1alpha1.ModbusDevicePropertyTypeUint:
		b, err = visitor.GetEndianness().ConvertUint32String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.ModbusDevicePropertyTypeUint64:
		b, err = visitor.GetEndianness().ConvertUint64String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.ModbusDevicePropertyTypeFloat32, v1alpha1.ModbusDevicePropertyTypeFloat:
		b, err = visitor.GetEndianness().ConvertFloat32String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.ModbusDevicePropertyTypeFloat64, v1alpha1.ModbusDevicePropertyTypeDouble:
		b, err = visitor.GetEndianness().ConvertFloat64String(prop.Value)
		if err != nil {
			return
		}
	case v1alpha1.ModbusDevicePropertyTypeBoolean:
		b, err = visitor.GetEndianness().ConvertBooleanString(prop.Value)
		if err != nil {
			return nil, err
		}
	case v1alpha1.ModbusDevicePropertyTypeHexString:
		b, err = hex.DecodeString(prop.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the hex value to bytes")
		}
	case v1alpha1.ModbusDevicePropertyTypeBinaryString:
		b, err = converter.DecodeBinaryString(prop.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the binary value to bytes")
		}
	case v1alpha1.ModbusDevicePropertyTypeBase64String:
		b, err = converter.DecodeBase64String(prop.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the base64 value to bytes")
		}
	case v1alpha1.ModbusDevicePropertyTypeString:
		b, err = visitor.GetEndianness().ConvertString(prop.Value)
		if err != nil {
			return
		}

		// NB(thxCode) 1 quantity == 16 bits = 2 bytes
		var bSizeExpected = int(visitor.Quantity) * 2
		var bSizeActual = len(b)
		if bSizeActual >= bSizeExpected {
			b = b[:bSizeExpected]
		} else {
			// pads
			var padding = make([]byte, bSizeExpected)
			copy(padding, b)
			b = padding
		}
	default:
		return nil, errors.Errorf("cannot convert the %s value to bytes", prop.Type)
	}

	return
}

// parseValueFromBytes parses the property value from a byte array according to the property type,
// and returns the operation result too.
func parseValueFromBytes(data []byte, prop *v1alpha1.ModbusDeviceProperty) (value string, result string, err error) {
	var visitor = &prop.Visitor

	switch prop.Type {
	case v1alpha1.ModbusDevicePropertyTypeInt8, v1alpha1.ModbusDevicePropertyTypeInt16:
		// NB(thxCode) in modbus, we use 16bits to store 8bits numeric data.
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
	case v1alpha1.ModbusDevicePropertyTypeInt32, v1alpha1.ModbusDevicePropertyTypeInt:
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
	case v1alpha1.ModbusDevicePropertyTypeInt64:
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
	case v1alpha1.ModbusDevicePropertyTypeUint8, v1alpha1.ModbusDevicePropertyTypeUint16:
		// NB(thxCode) in modbus, we use 16bits to store 8bits numeric data.
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
	case v1alpha1.ModbusDevicePropertyTypeUint32, v1alpha1.ModbusDevicePropertyTypeUint:
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
	case v1alpha1.ModbusDevicePropertyTypeUint64:
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
	case v1alpha1.ModbusDevicePropertyTypeFloat32, v1alpha1.ModbusDevicePropertyTypeFloat:
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
	case v1alpha1.ModbusDevicePropertyTypeFloat64, v1alpha1.ModbusDevicePropertyTypeDouble:
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
	case v1alpha1.ModbusDevicePropertyTypeBoolean:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseBooleanStringValue(data)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.ModbusDevicePropertyTypeHexString:
		value = hex.EncodeToString(data)
	case v1alpha1.ModbusDevicePropertyTypeBinaryString:
		value = converter.EncodeBinaryToString(data)
	case v1alpha1.ModbusDevicePropertyTypeBase64String:
		value = converter.EncodeBase64ToString(data)
	case v1alpha1.ModbusDevicePropertyTypeString:
		value, err = visitor.GetEndianness().ParseString(data)
		if err != nil {
			return
		}
	default:
		err = errors.Errorf("cannot parse bytes to %s type", prop.Type)
	}

	return
}

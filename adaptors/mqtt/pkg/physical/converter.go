package physical

import (
	"encoding/hex"
	"strconv"

	"github.com/pkg/errors"

	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
	endianapi "github.com/rancher/octopus/pkg/endian/api"
	"github.com/rancher/octopus/pkg/util/converter"
)

// doArithmeticOperations helps to calculate the raw value with operations.
func doArithmeticOperations(raw interface{}, visitor *v1alpha1.MQTTDevicePropertyVisitor) (string, error) {
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
		case v1alpha1.MQTTDevicePropertyValueArithmeticAdd:
			result = result + operationValue
		case v1alpha1.MQTTDevicePropertyValueArithmeticSubtract:
			result = result - operationValue
		case v1alpha1.MQTTDevicePropertyValueArithmeticMultiply:
			result = result * operationValue
		case v1alpha1.MQTTDevicePropertyValueArithmeticDivide:
			result = result / operationValue
		}
	}
	return strconv.FormatFloat(result, 'f', visitor.GetArithmeticOperationPrecision(), 64), nil
}

// convertBytesContentTypeValueToBytes converts the property value of bytes content type to a MQTT payload according to the property type.
func convertBytesContentTypeValueToBytes(prop *v1alpha1.MQTTDeviceProperty) (b []byte, err error) {
	var visitor = &prop.Visitor

	switch prop.Type {
	case v1alpha1.MQTTDevicePropertyTypeInt8:
		b, err = visitor.GetEndianness().ConvertInt8String(prop.Value.String())
		if err != nil {
			return
		}
	case v1alpha1.MQTTDevicePropertyTypeInt16:
		b, err = visitor.GetEndianness().ConvertInt16String(prop.Value.String())
		if err != nil {
			return
		}
	case v1alpha1.MQTTDevicePropertyTypeInt32, v1alpha1.MQTTDevicePropertyTypeInt:
		b, err = visitor.GetEndianness().ConvertInt32String(prop.Value.String())
		if err != nil {
			return
		}
	case v1alpha1.MQTTDevicePropertyTypeInt64:
		b, err = visitor.GetEndianness().ConvertInt64String(prop.Value.String())
		if err != nil {
			return
		}
	case v1alpha1.MQTTDevicePropertyTypeUint8:
		b, err = visitor.GetEndianness().ConvertUint8String(prop.Value.String())
		if err != nil {
			return
		}
	case v1alpha1.MQTTDevicePropertyTypeUint16:
		b, err = visitor.GetEndianness().ConvertUint16String(prop.Value.String())
		if err != nil {
			return
		}
	case v1alpha1.MQTTDevicePropertyTypeUint32, v1alpha1.MQTTDevicePropertyTypeUint:
		b, err = visitor.GetEndianness().ConvertUint32String(prop.Value.String())
		if err != nil {
			return
		}
	case v1alpha1.MQTTDevicePropertyTypeUint64:
		b, err = visitor.GetEndianness().ConvertUint64String(prop.Value.String())
		if err != nil {
			return
		}
	case v1alpha1.MQTTDevicePropertyTypeFloat32, v1alpha1.MQTTDevicePropertyTypeFloat:
		b, err = visitor.GetEndianness().ConvertFloat32String(prop.Value.String())
		if err != nil {
			return
		}
	case v1alpha1.MQTTDevicePropertyTypeFloat64, v1alpha1.MQTTDevicePropertyTypeDouble:
		b, err = visitor.GetEndianness().ConvertFloat64String(prop.Value.String())
		if err != nil {
			return
		}
	case v1alpha1.MQTTDevicePropertyTypeBoolean:
		b, err = visitor.GetEndianness().ConvertBooleanString(prop.Value.String())
		if err != nil {
			return
		}
	case v1alpha1.MQTTDevicePropertyTypeHexString:
		b, err = hex.DecodeString(prop.Value.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the hex value to bytes")
		}
	case v1alpha1.MQTTDevicePropertyTypeBinaryString:
		b, err = converter.DecodeBinaryString(prop.Value.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the binary value to bytes")
		}
	case v1alpha1.MQTTDevicePropertyTypeBase64String:
		b, err = converter.DecodeBase64String(prop.Value.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the base64 value to bytes")
		}
	case v1alpha1.MQTTDevicePropertyTypeString:
		b, err = visitor.GetEndianness().ConvertString(prop.Value.String())
		if err != nil {
			return
		}
	default:
		return nil, errors.Errorf("cannot convert the %s value to bytes", prop.Type)
	}

	return
}

// parseTextContentTypeValueFromBytes parses the property value of text content type from a byte array according to the property type,
// // and returns the operation result too.
func parseTextContentTypeValueFromBytes(data []byte, prop *v1alpha1.MQTTDeviceProperty) (value string, operationResult string, err error) {
	var visitor = &prop.Visitor

	if data[0] == '"' && data[len(data)-1] == '"' {
		value = converter.UnsafeBytesToString(data[1 : len(data)-1])
	} else {
		value = converter.UnsafeBytesToString(data)
	}

	// executes arithmetic operations
	switch prop.Type {
	case v1alpha1.MQTTDevicePropertyTypeInt8,
		v1alpha1.MQTTDevicePropertyTypeInt16,
		v1alpha1.MQTTDevicePropertyTypeInt32, v1alpha1.MQTTDevicePropertyTypeInt,
		v1alpha1.MQTTDevicePropertyTypeInt64,
		v1alpha1.MQTTDevicePropertyTypeUint8,
		v1alpha1.MQTTDevicePropertyTypeUint16,
		v1alpha1.MQTTDevicePropertyTypeUint32, v1alpha1.MQTTDevicePropertyTypeUint,
		v1alpha1.MQTTDevicePropertyTypeUint64,
		v1alpha1.MQTTDevicePropertyTypeFloat32, v1alpha1.MQTTDevicePropertyTypeFloat,
		v1alpha1.MQTTDevicePropertyTypeFloat64, v1alpha1.MQTTDevicePropertyTypeDouble:
		operationResult, err = doArithmeticOperations(value, visitor)
	}

	return
}

// parseBytesContentTypeValueFromBytes parses the property value of bytes content type from a byte array according to the property type,
// and returns the operation result too.
func parseBytesContentTypeValueFromBytes(data []byte, prop *v1alpha1.MQTTDeviceProperty) (value string, result string, err error) {
	var visitor = &prop.Visitor

	switch prop.Type {
	case v1alpha1.MQTTDevicePropertyTypeInt8:
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
	case v1alpha1.MQTTDevicePropertyTypeInt16:
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
	case v1alpha1.MQTTDevicePropertyTypeInt32, v1alpha1.MQTTDevicePropertyTypeInt:
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
	case v1alpha1.MQTTDevicePropertyTypeInt64:
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
	case v1alpha1.MQTTDevicePropertyTypeUint8:
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
	case v1alpha1.MQTTDevicePropertyTypeUint16:
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
	case v1alpha1.MQTTDevicePropertyTypeUint32, v1alpha1.MQTTDevicePropertyTypeUint:
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
	case v1alpha1.MQTTDevicePropertyTypeUint64:
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
	case v1alpha1.MQTTDevicePropertyTypeFloat32, v1alpha1.MQTTDevicePropertyTypeFloat:
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
	case v1alpha1.MQTTDevicePropertyTypeFloat64, v1alpha1.MQTTDevicePropertyTypeDouble:
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
	case v1alpha1.MQTTDevicePropertyTypeBoolean:
		var sv endianapi.StringValue
		sv, err = visitor.GetEndianness().ParseBooleanStringValue(data)
		if err != nil {
			return
		}
		value = sv.String()
	case v1alpha1.MQTTDevicePropertyTypeHexString:
		value = hex.EncodeToString(data)
	case v1alpha1.MQTTDevicePropertyTypeBinaryString:
		value = converter.EncodeBinaryToString(data)
	case v1alpha1.MQTTDevicePropertyTypeBase64String:
		value = converter.EncodeBase64ToString(data)
	case v1alpha1.MQTTDevicePropertyTypeString:
		value, err = visitor.GetEndianness().ParseString(data)
		if err != nil {
			return
		}
	default:
		err = errors.Errorf("cannot parse bytes to %s type", prop.Type)
	}

	return
}

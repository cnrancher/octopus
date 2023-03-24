package physical

import (
	"encoding/hex"
	"strconv"

	"github.com/gopcua/opcua/ua"
	"github.com/pkg/errors"

	"github.com/rancher/octopus/adaptors/opcua/api/v1alpha1"
	"github.com/rancher/octopus/pkg/util/converter"
)

// doArithmeticOperations helps to calculate the raw value with operations,
// and returns the calculated raw result in 6 digit precision.
func doArithmeticOperations(raw float64, visitor *v1alpha1.OPCUADevicePropertyVisitor) (string, error) {
	if len(visitor.ArithmeticOperations) == 0 {
		return "", nil
	}

	var result = raw
	for _, executeOperation := range visitor.ArithmeticOperations {
		operationValue, err := strconv.ParseFloat(executeOperation.Value, 64)
		if err != nil {
			return "", errors.Wrapf(err, "failed to parse %s operation's value", executeOperation.Type)
		}
		switch executeOperation.Type {
		case v1alpha1.OPCUADevicePropertyValueArithmeticAdd:
			result = result + operationValue
		case v1alpha1.OPCUADevicePropertyValueArithmeticSubtract:
			result = result - operationValue
		case v1alpha1.OPCUADevicePropertyValueArithmeticMultiply:
			result = result * operationValue
		case v1alpha1.OPCUADevicePropertyValueArithmeticDivide:
			result = result / operationValue
		}
	}
	return strconv.FormatFloat(result, 'g', visitor.GetArithmeticOperationPrecision(), 64), nil
}

// convertValueToVariant converts the property value to a UA variant according to the property type.
func convertValueToVariant(prop *v1alpha1.OPCUADeviceProperty) (*ua.Variant, error) {
	var data *ua.Variant

	switch prop.Type {
	case v1alpha1.OPCUADevicePropertyTypeInt8:
		var val, err = strconv.ParseInt(prop.Value, 10, 8)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to int8")
		}
		data, err = ua.NewVariant(int8(val))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with int8 value")
		}
	case v1alpha1.OPCUADevicePropertyTypeInt16:
		var val, err = strconv.ParseInt(prop.Value, 10, 16)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to int16")
		}
		data, err = ua.NewVariant(int16(val))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with int16 value")
		}
	case v1alpha1.OPCUADevicePropertyTypeInt32, v1alpha1.OPCUADevicePropertyTypeInt:
		var val, err = strconv.ParseInt(prop.Value, 10, 32)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to int32/int")
		}
		data, err = ua.NewVariant(int32(val))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with int32 value")
		}
	case v1alpha1.OPCUADevicePropertyTypeInt64:
		var val, err = strconv.ParseInt(prop.Value, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to int64")
		}
		data, err = ua.NewVariant(val)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with int64 value")
		}
	case v1alpha1.OPCUADevicePropertyTypeUint8:
		var val, err = strconv.ParseUint(prop.Value, 10, 8)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to uint8")
		}
		data, err = ua.NewVariant(uint8(val))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with uint8 value")
		}
	case v1alpha1.OPCUADevicePropertyTypeUint16:
		var val, err = strconv.ParseUint(prop.Value, 10, 16)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to uint16")
		}
		data, err = ua.NewVariant(uint16(val))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with uint16 value")
		}
	case v1alpha1.OPCUADevicePropertyTypeUint32, v1alpha1.OPCUADevicePropertyTypeUint:
		var val, err = strconv.ParseUint(prop.Value, 10, 32)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to uint32/uint")
		}
		data, err = ua.NewVariant(uint32(val))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with uint32 value")
		}
	case v1alpha1.OPCUADevicePropertyTypeUint64:
		var val, err = strconv.ParseUint(prop.Value, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to uint64")
		}
		data, err = ua.NewVariant(val)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with uint64 value")
		}
	case v1alpha1.OPCUADevicePropertyTypeFloat32, v1alpha1.OPCUADevicePropertyTypeFloat:
		var val, err = strconv.ParseFloat(prop.Value, 32)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to float32/float")
		}
		data, err = ua.NewVariant(float32(val))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with float32 value")
		}
	case v1alpha1.OPCUADevicePropertyTypeFloat64, v1alpha1.OPCUADevicePropertyTypeDouble:
		var val, err = strconv.ParseFloat(prop.Value, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to float64/double")
		}
		data, err = ua.NewVariant(val)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with float64 value")
		}
	case v1alpha1.OPCUADevicePropertyTypeBoolean:
		var val, err = strconv.ParseBool(prop.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the value to boolean")
		}
		data, err = ua.NewVariant(val)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with boolean value")
		}
	case v1alpha1.OPCUADevicePropertyTypeHexString:
		var val, err = hex.DecodeString(prop.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the hex value to bytes")
		}
		data, err = ua.NewVariant(val)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with byteString value")
		}
	case v1alpha1.OPCUADevicePropertyTypeBinaryString:
		var val, err = converter.DecodeBinaryString(prop.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the binary value to bytes")
		}
		data, err = ua.NewVariant(val)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with byteString value")
		}
	case v1alpha1.OPCUADevicePropertyTypeBase64String:
		var val, err = converter.DecodeBase64String(prop.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the base64 value to bytes")
		}
		data, err = ua.NewVariant(val)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with byteString value")
		}
	case v1alpha1.OPCUADevicePropertyTypeString:
		var err error
		data, err = ua.NewVariant(prop.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create UA variant with string value")
		}
	default:
		return nil, errors.Errorf("cannot convert the %s value to UA variant", prop.Type)
	}

	return data, nil
}

// parseValueFromBytes parses the property value from a UA variant according to the property type,
// and returns the operation result too.
func parseValueFromVariant(data *ua.Variant, prop *v1alpha1.OPCUADeviceProperty) (value string, operationResult string, err error) {
	var visitor = &prop.Visitor

	switch prop.Type {
	case v1alpha1.OPCUADevicePropertyTypeInt8:
		if data.Type() != ua.TypeIDSByte {
			err = errors.Errorf("failed to parse UA variant to int8 as the actual type is %s", data.Type())
			return
		}
		var val = int8(data.Int())
		value = strconv.FormatInt(int64(val), 10)
		operationResult, err = doArithmeticOperations(float64(val), visitor)
		if err != nil {
			return
		}
	case v1alpha1.OPCUADevicePropertyTypeInt16:
		if data.Type() != ua.TypeIDInt16 {
			err = errors.Errorf("failed to parse UA variant to int16 as the actual type is %s", data.Type())
			return
		}
		var val = int16(data.Int())
		value = strconv.FormatInt(int64(val), 10)
		operationResult, err = doArithmeticOperations(float64(val), visitor)
		if err != nil {
			return
		}
	case v1alpha1.OPCUADevicePropertyTypeInt32, v1alpha1.OPCUADevicePropertyTypeInt:
		if data.Type() != ua.TypeIDInt32 {
			err = errors.Errorf("failed to parse UA variant to int32 as the actual type is %s", data.Type())
			return
		}
		var val = int32(data.Int())
		value = strconv.FormatInt(int64(val), 10)
		operationResult, err = doArithmeticOperations(float64(val), visitor)
		if err != nil {
			return
		}
	case v1alpha1.OPCUADevicePropertyTypeInt64:
		if data.Type() != ua.TypeIDInt64 {
			err = errors.Errorf("failed to parse UA variant to int64 as the actual type is %s", data.Type())
			return
		}
		var val = data.Int()
		value = strconv.FormatInt(val, 10)
		operationResult, err = doArithmeticOperations(float64(val), visitor)
		if err != nil {
			return
		}
	case v1alpha1.OPCUADevicePropertyTypeUint8:
		if data.Type() != ua.TypeIDByte {
			err = errors.Errorf("failed to parse UA variant to uint8 as the actual type is %s", data.Type())
			return
		}
		var val = uint8(data.Uint())
		value = strconv.FormatUint(uint64(val), 10)
		operationResult, err = doArithmeticOperations(float64(val), visitor)
		if err != nil {
			return
		}
	case v1alpha1.OPCUADevicePropertyTypeUint16:
		if data.Type() != ua.TypeIDUint16 {
			err = errors.Errorf("failed to parse UA variant to uint16 as the actual type is %s", data.Type())
			return
		}
		var val = uint16(data.Uint())
		value = strconv.FormatUint(uint64(val), 10)
		operationResult, err = doArithmeticOperations(float64(val), visitor)
		if err != nil {
			return
		}
	case v1alpha1.OPCUADevicePropertyTypeUint32, v1alpha1.OPCUADevicePropertyTypeUint:
		if data.Type() != ua.TypeIDUint32 {
			err = errors.Errorf("failed to parse UA variant to uint32 as the actual type is %s", data.Type())
			return
		}
		var val = uint32(data.Uint())
		value = strconv.FormatUint(uint64(val), 10)
		operationResult, err = doArithmeticOperations(float64(val), visitor)
		if err != nil {
			return
		}
	case v1alpha1.OPCUADevicePropertyTypeUint64:
		if data.Type() != ua.TypeIDUint64 {
			err = errors.Errorf("failed to parse UA variant to uint64 as the actual type is %s", data.Type())
			return
		}
		var val = data.Uint()
		value = strconv.FormatUint(val, 10)
		operationResult, err = doArithmeticOperations(float64(val), visitor)
		if err != nil {
			return
		}
	case v1alpha1.OPCUADevicePropertyTypeFloat32, v1alpha1.OPCUADevicePropertyTypeFloat:
		if data.Type() != ua.TypeIDFloat {
			err = errors.Errorf("failed to parse UA variant to float32 as the actual type is %s", data.Type())
			return
		}
		var val = data.Float()
		value = strconv.FormatFloat(val, 'g', -1, 32)
		operationResult, err = doArithmeticOperations(val, visitor)
		if err != nil {
			return
		}
	case v1alpha1.OPCUADevicePropertyTypeFloat64, v1alpha1.OPCUADevicePropertyTypeDouble:
		if data.Type() != ua.TypeIDDouble {
			err = errors.Errorf("failed to parse UA variant to float64 as the actual type is %s", data.Type())
			return
		}
		var val = data.Float()
		value = strconv.FormatFloat(val, 'g', -1, 64)
		operationResult, err = doArithmeticOperations(val, visitor)
		if err != nil {
			return
		}
	case v1alpha1.OPCUADevicePropertyTypeBoolean:
		if data.Type() != ua.TypeIDBoolean {
			err = errors.Errorf("failed to parse UA variant to boolean as the actual type is %s", data.Type())
			return
		}
		var val = data.Bool()
		value = strconv.FormatBool(val)
	case v1alpha1.OPCUADevicePropertyTypeHexString:
		if data.Type() != ua.TypeIDByteString {
			err = errors.Errorf("failed to parse UA variant to bytes as the actual type is %s", data.Type())
			return
		}
		var val = data.ByteString()
		value = hex.EncodeToString(val)
	case v1alpha1.OPCUADevicePropertyTypeBinaryString:
		if data.Type() != ua.TypeIDByteString {
			err = errors.Errorf("failed to parse UA variant to bytes as the actual type is %s", data.Type())
			return
		}
		var val = data.ByteString()
		value = converter.EncodeBinaryToString(val)
	case v1alpha1.OPCUADevicePropertyTypeBase64String:
		if data.Type() != ua.TypeIDByteString {
			err = errors.Errorf("failed to parse UA variant to bytes as the actual type is %s", data.Type())
			return
		}
		var val = data.ByteString()
		value = converter.EncodeBase64ToString(val)
	case v1alpha1.OPCUADevicePropertyTypeString:
		if data.Type() != ua.TypeIDString {
			err = errors.Errorf("failed to parse UA variant to string as the actual type is %s", data.Type())
			return
		}
		value = data.String()
	default:
		err = errors.Errorf("cannot parse UA variant to %s type", prop.Type)
	}

	return
}

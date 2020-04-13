package physical

import (
	"errors"
	"strconv"

	"github.com/gopcua/opcua/ua"
	"github.com/rancher/octopus/adaptors/opcua/api/v1alpha1"
)

var typeMap = map[ua.TypeID]v1alpha1.PropertyDataType{
	ua.TypeIDInt16:   v1alpha1.PropertyDataTypeInt16,
	ua.TypeIDInt32:   v1alpha1.PropertyDataTypeInt32,
	ua.TypeIDInt64:   v1alpha1.PropertyDataTypeInt64,
	ua.TypeIDUint16:  v1alpha1.PropertyDataTypeUInt16,
	ua.TypeIDUint32:  v1alpha1.PropertyDataTypeUInt32,
	ua.TypeIDUint64:  v1alpha1.PropertyDataTypeUInt64,
	ua.TypeIDDouble:  v1alpha1.PropertyDataTypeDouble,
	ua.TypeIDFloat:   v1alpha1.PropertyDataTypeFloat,
	ua.TypeIDBoolean: v1alpha1.PropertyDataTypeBoolean,
	ua.TypeIDString:  v1alpha1.PropertyDataTypeString,
}

func VariantToString(dataType ua.TypeID, input *ua.Variant) string {
	switch dataType {
	case ua.TypeIDBoolean:
		return strconv.FormatBool(input.Bool())
	case ua.TypeIDFloat:
		return strconv.FormatFloat(input.Float(), 'g', -1, 32)
	case ua.TypeIDDouble:
		return strconv.FormatFloat(input.Float(), 'g', -1, 64)
	case ua.TypeIDInt16, ua.TypeIDInt32, ua.TypeIDInt64:
		return strconv.FormatInt(input.Int(), 10)
	case ua.TypeIDUint16, ua.TypeIDUint32, ua.TypeIDUint64:
		return strconv.FormatUint(input.Uint(), 10)
	case ua.TypeIDString:
		return input.String()
	}
	return ""
}

func StringToVariant(dataType v1alpha1.PropertyDataType, input string) (*ua.Variant, error) {
	switch dataType {
	case v1alpha1.PropertyDataTypeString:
		return ua.NewVariant(input)
	case v1alpha1.PropertyDataTypeInt64:
		result, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(result)
	case v1alpha1.PropertyDataTypeInt32:
		result, err := strconv.ParseInt(input, 10, 32)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(result)
	case v1alpha1.PropertyDataTypeInt16:
		result, err := strconv.ParseInt(input, 10, 16)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(result)
	case v1alpha1.PropertyDataTypeUInt64:
		result, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(result)
	case v1alpha1.PropertyDataTypeUInt32:
		result, err := strconv.ParseUint(input, 10, 32)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(result)
	case v1alpha1.PropertyDataTypeUInt16:
		result, err := strconv.ParseUint(input, 10, 16)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(result)
	case v1alpha1.PropertyDataTypeFloat:
		result, err := strconv.ParseFloat(input, 32)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(result)
	case v1alpha1.PropertyDataTypeDouble:
		result, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(result)
	case v1alpha1.PropertyDataTypeBoolean:
		result, err := strconv.ParseBool(input)
		if err != nil {
			return nil, err
		}
		return ua.NewVariant(result)
	}
	return nil, errors.New("invalid data type")
}

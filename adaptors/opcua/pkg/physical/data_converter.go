package physical

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gopcua/opcua/ua"
	"github.com/rancher/octopus/adaptors/opcua/api/v1alpha1"
)

var typeMap = map[ua.TypeID]v1alpha1.OPCUADevicePropertyType{
	ua.TypeIDInt16:      v1alpha1.OPCUADevicePropertyTypeInt16,
	ua.TypeIDInt32:      v1alpha1.OPCUADevicePropertyTypeInt32,
	ua.TypeIDInt64:      v1alpha1.OPCUADevicePropertyTypeInt64,
	ua.TypeIDUint16:     v1alpha1.OPCUADevicePropertyTypeUInt16,
	ua.TypeIDUint32:     v1alpha1.OPCUADevicePropertyTypeUInt32,
	ua.TypeIDUint64:     v1alpha1.OPCUADevicePropertyTypeUInt64,
	ua.TypeIDDouble:     v1alpha1.OPCUADevicePropertyTypeDouble,
	ua.TypeIDFloat:      v1alpha1.OPCUADevicePropertyTypeFloat,
	ua.TypeIDBoolean:    v1alpha1.OPCUADevicePropertyTypeBoolean,
	ua.TypeIDString:     v1alpha1.OPCUADevicePropertyTypeString,
	ua.TypeIDByteString: v1alpha1.OPCUADevicePropertyTypeByteString,
	ua.TypeIDDateTime:   v1alpha1.OPCUADevicePropertyTypeDatetime,
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
	case ua.TypeIDByteString:
		return string(input.ByteString())
	default:
		return fmt.Sprintf("%v", input.Value())
	}
}

func StringToVariant(dataType v1alpha1.OPCUADevicePropertyType, input string) (*ua.Variant, error) {
	var result interface{}
	var err error
	switch dataType {
	case v1alpha1.OPCUADevicePropertyTypeString:
		return ua.NewVariant(input)
	case v1alpha1.OPCUADevicePropertyTypeInt64:
		result, err = strconv.ParseInt(input, 10, 64)
		if err != nil {
			return nil, err
		}
	case v1alpha1.OPCUADevicePropertyTypeInt32:
		parsed, err := strconv.ParseInt(input, 10, 32)
		if err != nil {
			return nil, err
		}
		result = int32(parsed)
	case v1alpha1.OPCUADevicePropertyTypeInt16:
		parsed, err := strconv.ParseInt(input, 10, 16)
		if err != nil {
			return nil, err
		}
		result = int16(parsed)
	case v1alpha1.OPCUADevicePropertyTypeUInt64:
		result, err = strconv.ParseUint(input, 10, 64)
		if err != nil {
			return nil, err
		}
	case v1alpha1.OPCUADevicePropertyTypeUInt32:
		parsed, err := strconv.ParseUint(input, 10, 32)
		if err != nil {
			return nil, err
		}
		result = uint32(parsed)
	case v1alpha1.OPCUADevicePropertyTypeUInt16:
		parsed, err := strconv.ParseUint(input, 10, 16)
		if err != nil {
			return nil, err
		}
		result = uint16(parsed)
	case v1alpha1.OPCUADevicePropertyTypeFloat:
		parsed, err := strconv.ParseFloat(input, 32)
		if err != nil {
			return nil, err
		}
		result = float32(parsed)
	case v1alpha1.OPCUADevicePropertyTypeDouble:
		result, err = strconv.ParseFloat(input, 64)
		if err != nil {
			return nil, err
		}
	case v1alpha1.OPCUADevicePropertyTypeBoolean:
		result, err = strconv.ParseBool(input)
		if err != nil {
			return nil, err
		}
	case v1alpha1.OPCUADevicePropertyTypeDatetime:
		result, err = time.Parse("2020-06-03T06:40:21.268109", input)
		if err != nil {
			return nil, err
		}
	case v1alpha1.OPCUADevicePropertyTypeByteString:
		result = []byte(input)
	default:
		return nil, errors.New("invalid data type")
	}
	return ua.NewVariant(result)
}

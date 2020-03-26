package physical

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
)

// convert read data to string value
func ByteArrayToString(input []byte, dataType v1alpha1.PropertyDataType) (string, error) {
	var result string
	switch dataType {
	case v1alpha1.PropertyDataTypeString:
		result = string(input)
	case v1alpha1.PropertyDataTypeFloat:
		arr, err := toTargetLength(input, 8)
		if err != nil {
			return "", err
		}
		value := binary.BigEndian.Uint64(arr)
		result = fmt.Sprint(math.Float64frombits(value))
	case v1alpha1.PropertyDataTypeInt:
		arr, err := toTargetLength(input, 8)
		if err != nil {
			return "", err
		}
		value := binary.BigEndian.Uint64(arr)
		result = strconv.Itoa(int(value))
	case v1alpha1.PropertyDataTypeBoolean:
		b := input[len(input)-1]
		if b == 0 {
			result = "false"
		} else if b == 1 {
			result = "true"
		} else {
			return "", errors.New("invalid boolean value")
		}
	default:
		return "", errors.New("invalid data type")
	}
	return result, nil
}

// convert written data to byte array according to datatype
func StringToByteArray(input string, dataType v1alpha1.PropertyDataType, length int) ([]byte, error) {
	var data []byte
	switch dataType {
	case v1alpha1.PropertyDataTypeString:
		data = []byte(input)
	case v1alpha1.PropertyDataTypeBoolean:
		b, err := strconv.ParseBool(input)
		if err != nil {
			return nil, err
		}
		if b == true {
			data = []byte{1}
		} else {
			data = []byte{0}
		}
	case v1alpha1.PropertyDataTypeInt:
		data = make([]byte, 8)
		i, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return nil, err
		}
		binary.BigEndian.PutUint64(data, i)
	case v1alpha1.PropertyDataTypeFloat:
		data = make([]byte, 8)
		f, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return nil, err
		}
		binary.BigEndian.PutUint64(data, math.Float64bits(f))
	default:
		return nil, errors.New("invalid data type")
	}
	return toTargetLength(data, length)
}

// Pad or trim byte array to target length
// Short input gets zeros padded to the left, long input gets left bits trimmed
func toTargetLength(input []byte, length int) ([]byte, error) {
	l := len(input)
	if l == length {
		return input, nil
	}
	if l > length {
		if input[l-length-1] != 0 {
			return nil, errors.New("input is longer than target length")
		}
		return input[l-length:], nil
	}
	tmp := make([]byte, length)
	copy(tmp[length-l:], input)
	return tmp, nil
}

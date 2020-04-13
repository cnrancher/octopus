package physical

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
	"github.com/sirupsen/logrus"
)

// convert read data to string value
func ByteArrayToString(input []byte, dataType v1alpha1.PropertyDataType, operations []v1alpha1.ModbusOperations) (string, error) {
	var result string
	switch dataType {
	case v1alpha1.PropertyDataTypeString:
		result = string(input)
	case v1alpha1.PropertyDataTypeInt, v1alpha1.PropertyDataTypeFloat:
		arr, err := toTargetLength(input, 8)
		if err != nil {
			return "", err
		}
		value := binary.BigEndian.Uint64(arr)
		converted := convertReadData(float64(value), operations)
		result = fmt.Sprint(converted)
		if dataType == v1alpha1.PropertyDataTypeInt {
			result = fmt.Sprint(int(converted))
		}
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
	case v1alpha1.PropertyDataTypeInt, v1alpha1.PropertyDataTypeFloat:
		data = make([]byte, 8)
		i, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return nil, err
		}
		binary.BigEndian.PutUint64(data, i)
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

// ConvertReadData helps to convert the number read from the device into meaningful data
func convertReadData(result float64, operations []v1alpha1.ModbusOperations) float64 {
	for _, executeOperation := range operations {
		operationValue, err := strconv.ParseFloat(executeOperation.OperationValue, 64)
		if err != nil {
			logrus.Error(err, "failed to parse operation value")
		}
		switch executeOperation.OperationType {
		case v1alpha1.OperationAdd:
			result = result + operationValue
		case v1alpha1.OperationSubtract:
			result = result - operationValue
		case v1alpha1.OperationMultiply:
			result = result * operationValue
		case v1alpha1.OperationDivide:
			result = result / operationValue
		}
	}
	return result
}

package physical

import (
	"strconv"

	"github.com/rancher/octopus/adaptors/ble/api/v1alpha1"
	"github.com/sirupsen/logrus"
)

// ConvertReadData helps to convert the data read from the device into meaningful data
func ConvertReadData(dataConverter v1alpha1.BluetoothDataConverter, data []byte) float64 {
	var intermediateResult uint64
	var initialValue []byte
	var initialStringValue = ""
	if dataConverter.StartIndex <= dataConverter.EndIndex {
		for index := dataConverter.StartIndex; index <= dataConverter.EndIndex; index++ {
			initialValue = append(initialValue, data[index])
		}
	} else {
		for index := dataConverter.StartIndex; index >= dataConverter.EndIndex; index-- {
			initialValue = append(initialValue, data[index])
		}
	}
	for _, value := range initialValue {
		initialStringValue = initialStringValue + strconv.Itoa(int(value))
	}
	initialByteValue, _ := strconv.ParseUint(initialStringValue, 16, 16)

	if dataConverter.ShiftLeft != 0 {
		intermediateResult = initialByteValue << dataConverter.ShiftLeft
	} else if dataConverter.ShiftRight != 0 {
		intermediateResult = initialByteValue >> dataConverter.ShiftRight
	}
	finalResult := float64(intermediateResult)
	for _, executeOperation := range dataConverter.OrderOfOperations {
		operationValue, err := strconv.ParseFloat(executeOperation.OperationValue, 64)
		if err != nil {
			logrus.Error(err, "failed to parse operation value")
		}
		switch executeOperation.OperationType {
		case v1alpha1.OperationAdd:
			finalResult = finalResult + operationValue
		case v1alpha1.OperationSubtract:
			finalResult = finalResult - operationValue
		case v1alpha1.OperationMultiply:
			finalResult = finalResult * operationValue
		case v1alpha1.OperationDivide:
			finalResult = finalResult / operationValue
		}
	}
	return finalResult
}

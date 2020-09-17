package physical

import (
	"strconv"

	"github.com/pkg/errors"

	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
)

// doArithmeticOperations helps to calculate the raw value with operations,
// and returns the calculated raw result in 6 digit precision.
func doArithmeticOperations(raw float64, operations []v1alpha1.ModbusDeviceArithmeticOperation) (string, error) {
	if len(operations) == 0 {
		return "", nil
	}

	var result = raw
	for _, executeOperation := range operations {
		operationValue, err := strconv.ParseFloat(executeOperation.Value, 64)
		if err != nil {
			return "", errors.Wrapf(err, "failed to parse %s operation's value", executeOperation.Type)
		}
		switch executeOperation.Type {
		case v1alpha1.ModbusDeviceArithmeticAdd:
			result = result + operationValue
		case v1alpha1.ModbusDeviceArithmeticSubtract:
			result = result - operationValue
		case v1alpha1.ModbusDeviceArithmeticMultiply:
			result = result * operationValue
		case v1alpha1.ModbusDeviceArithmeticDivide:
			result = result / operationValue
		}
	}
	return strconv.FormatFloat(result, byte('f'), 6, 64), nil
}

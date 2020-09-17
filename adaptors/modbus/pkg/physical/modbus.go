package physical

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"

	"github.com/pkg/errors"

	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
)

type (
	// registerWriteFunc specifies the func to write single/multiple register.
	registerWriteFunc func(address, quantity uint16, value []byte) (results []byte, err error)
	// registerReadFunc specifies the func to read single/multiple register.
	registerReadFunc func(address, quantity uint16) (results []byte, err error)
)

// write1BitRegister writes the given property's value to 1-bit register.
func write1BitRegister(prop *v1alpha1.ModbusDeviceProperty, write registerWriteFunc) error {
	if prop.Value == "" {
		return nil
	}

	var visitor = prop.Visitor

	if visitor.Quantity == 1 {
		// NB(thxCode) single quantity register can only be considered as boolean type.
		if prop.Type != v1alpha1.ModbusDevicePropertyTypeBoolean {
			return errors.Errorf("single 1-bit quantity %s can only set as boolean type", visitor.Register)
		}

		// parse value
		var data []byte
		switch prop.Value {
		case "true":
			data = []byte{1}
		case "false":
			data = []byte{0}
		default:
			return errors.Errorf("failed to convert the single 1-bit quantity %s's value to boolean", visitor.Register)
		}

		// write to single quantity register
		_, err := write(visitor.Offset, visitor.Quantity, data)
		if err != nil {
			return errors.Wrapf(err, "failed to write %s to single 1-bit quantity %s", prop.Value, visitor.Register)
		}
		return nil
	}

	// NB(thxCode) multiple quantities CoilRegister can only be considered as hexString type,
	// for example, if we write 20# - 29# CoilRegisters at the same time,
	// which means we can input "CD01" to indicate the following multiple CoilRegister values:
	// bit value  :  1  1  0  0  1  1  0  1  0  0  0  0  0  0  0  1
	// coil num   : 27 26 25 24 23 22 21 20  -  -  -  -  -  - 29 28
	// hex value  :  *  *  *  C| *  *  *  D| *  *  *  0| *  *  *  1|  <-- input/output
	// byte value :  *  *  *  *  *  *  *205| *  *  *  *  *  *  *  1|  <-- convert
	if prop.Type != v1alpha1.ModbusDevicePropertyTypeHexString {
		return errors.Errorf("multiple 1-bit quantities %s can only set as hexString type", visitor.Register)
	}

	// parse value
	if expectedHexStringSize := int((visitor.Quantity+8)/8) * 2; len(prop.Value) != expectedHexStringSize {
		return errors.Errorf("the length of multiple 1-bit quantities %s's hex value is invalid, expected is %d", visitor.Register, expectedHexStringSize)
	}
	var data, err = hex.DecodeString(prop.Value)
	if err != nil {
		return errors.Wrapf(err, "failed to convert the multiple 1-bit quantities %s's hex value to byte array", visitor.Register)
	}

	// write to multiple quantities register
	_, err = write(visitor.Offset, visitor.Quantity, data)
	if err != nil {
		return errors.Wrapf(err, "failed to write %s to multiple 1-bit quantities %s", prop.Value, visitor.Register)
	}
	return nil
}

// write16BitsRegister writes the given property's value to 16-bits register.
func write16BitsRegister(prop *v1alpha1.ModbusDeviceProperty, write registerWriteFunc) error {
	if prop.Value == "" {
		return nil
	}

	var visitor = prop.Visitor

	if visitor.Quantity == 1 {
		// parse value
		var data []byte
		switch prop.Type {
		case v1alpha1.ModbusDevicePropertyTypeHexString:
			// NB(thxCode) a 16-bits value can indicate by a size 4 hex string.
			if len(prop.Value) != 4 {
				return errors.Errorf("the length of single 16-bits quantity %s's hex value is invalid", visitor.Register)
			}

			var val, err = hex.DecodeString(prop.Value)
			if err != nil {
				return errors.Wrapf(err, "failed to convert the single 16-bits quantity %s's hex value to byte array", visitor.Register)
			}
			data = val
		case v1alpha1.ModbusDevicePropertyTypeInt16:
			var val, err = strconv.ParseInt(prop.Value, 10, 16)
			if err != nil {
				return errors.Wrapf(err, "failed to convert the single 16-bits quantity %s's value to int16", visitor.Register)
			}

			data = make([]byte, 2)
			visitor.Endianness.PutUint16(data, uint16(val))
		case v1alpha1.ModbusDevicePropertyTypeUint16:
			var val, err = strconv.ParseUint(prop.Value, 10, 16)
			if err != nil {
				return errors.Wrapf(err, "failed to convert the single 16-bits quantity %s's value to uint16", visitor.Register)
			}

			data = make([]byte, 2)
			visitor.Endianness.PutUint16(data, uint16(val))
		default:
			return errors.Errorf("single 16-bits quantity %s cannot set as %s type", visitor.Register, prop.Type)
		}

		// write to single quantity register
		_, err := write(visitor.Offset, visitor.Quantity, data)
		if err != nil {
			return errors.Wrapf(err, "failed to write %s to single 16-bits quantity %s", prop.Value, visitor.Register)
		}
		return nil
	}

	// parse value
	var data []byte
	switch prop.Type {
	case v1alpha1.ModbusDevicePropertyTypeInt, v1alpha1.ModbusDevicePropertyTypeInt32, v1alpha1.ModbusDevicePropertyTypeUint, v1alpha1.ModbusDevicePropertyTypeUint32:
		if visitor.Quantity != 2 {
			return errors.Errorf("multiple 16-bits quantities %s cannot set as int32 type", visitor.Register)
		}

		var val, err = strconv.ParseUint(prop.Value, 10, 32)
		if err != nil {
			return errors.Wrapf(err, "failed to convert the multiple 16-bits quantities %s's value to int32", visitor.Register)
		}

		data = make([]byte, 4)
		visitor.Endianness.PutUint32(data, uint32(val))
	case v1alpha1.ModbusDevicePropertyTypeInt64, v1alpha1.ModbusDevicePropertyTypeUint64:
		if visitor.Quantity != 4 {
			return errors.Errorf("multiple 16-bits quantities %s cannot as int64 type", visitor.Register)
		}

		var val, err = strconv.ParseUint(prop.Value, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "failed to convert the multiple 16-bits quantities %s's value to int64", visitor.Register)
		}

		data = make([]byte, 8)
		visitor.Endianness.PutUint64(data, val)
	case v1alpha1.ModbusDevicePropertyTypeFloat:
		if visitor.Quantity != 2 {
			return errors.Errorf("multiple 16-bits quantities %s cannot as float32 type", visitor.Register)
		}

		var val, err = strconv.ParseFloat(prop.Value, 32)
		if err != nil {
			return errors.Wrapf(err, "failed to convert the multiple 16-bits quantities %s's value to float32", visitor.Register)
		}

		data = make([]byte, 4)
		visitor.Endianness.PutUint32(data, math.Float32bits(float32(val)))
	case v1alpha1.ModbusDevicePropertyTypeDouble:
		if visitor.Quantity != 4 {
			return errors.Errorf("multiple 16-bits quantities %s cannot as float64 type", visitor.Register)
		}

		var val, err = strconv.ParseFloat(prop.Value, 64)
		if err != nil {
			return errors.Wrapf(err, "failed to convert the multiple 16-bits quantities %s's value to float64", visitor.Register)
		}

		data = make([]byte, 8)
		visitor.Endianness.PutUint64(data, math.Float64bits(val))
	default:
		return errors.Errorf("multiple 16-bits quantities %s cannot set as %s type", visitor.Register, prop.Type)
	}

	// write to multiple quantities register
	_, err := write(visitor.Offset, visitor.Quantity, data)
	if err != nil {
		return errors.Wrapf(err, "failed to write %s to multiple 16-bits quantities %s", prop.Value, visitor.Register)
	}
	return nil
}

// read1BitRegister reads the value of given property from 1-bit register.
func read1BitRegister(prop *v1alpha1.ModbusDeviceProperty, read registerReadFunc) (value string, operatedValue string, berr error) {
	var visitor = prop.Visitor

	if visitor.Quantity == 1 {
		// NB(thxCode) single quantity register can only be considered as boolean type.
		if prop.Type != v1alpha1.ModbusDevicePropertyTypeBoolean {
			return "", "", errors.Errorf("single 1-bit quantity %s can only set as boolean type", visitor.Register)
		}

		// read from single quantity register
		var val, err = read(visitor.Offset, visitor.Quantity)
		if err != nil {
			return "", "", errors.Wrapf(err, "failed to read from single 1-bit quantity %s", visitor.Register)
		}
		if len(val) != 1 {
			return "", "", errors.Errorf("failed to read from single 1-bit quantity %s, response bytes isn't in valid size", visitor.Register)
		}

		// parse data
		var data string
		switch val[0] {
		case 0:
			data = "false"
		default:
			data = "true"
		}
		return data, "", nil
	}

	// NB(thxCode) multiple quantities register can only be considered as hexString type.
	if prop.Type != v1alpha1.ModbusDevicePropertyTypeHexString {
		return "", "", errors.Errorf("multiple 1-bit quantities %s can only set as hexString type", visitor.Register)
	}

	// read from multiple quantities register
	var val, err = read(visitor.Offset, visitor.Quantity)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to read from multiple 1-bit quantities %s", visitor.Register)
	}
	if len(val) <= 1 {
		return "", "", errors.Errorf("failed to read from multiple 1-bits quantities %s, response bytes isn't in valid size", visitor.Register)
	}

	// parse data
	var data = hex.EncodeToString(val)
	return data, "", nil
}

// read16BitsRegister reads the value of given property from 16-bits register.
func read16BitsRegister(prop *v1alpha1.ModbusDeviceProperty, read registerReadFunc) (value string, operatedValue string, berr error) {
	var visitor = prop.Visitor

	if visitor.Quantity == 1 {
		// validate first
		switch prop.Type {
		case v1alpha1.ModbusDevicePropertyTypeHexString:
			// pass
		case v1alpha1.ModbusDevicePropertyTypeInt16, v1alpha1.ModbusDevicePropertyTypeUint16:
			// pass
		default:
			return "", "", errors.Errorf("single 16-bits quantity %s cannot set as %s type", visitor.Register, prop.Type)
		}

		// read from single quantity register
		var val, err = read(visitor.Offset, visitor.Quantity)
		if err != nil {
			return "", "", errors.Wrapf(err, "failed to read from single 16-bits quantity %s", visitor.Register)
		}
		if len(val) != 2 {
			return "", "", errors.Errorf("failed to read from single 16-bits quantity %s, response bytes isn't in valid size", visitor.Register)
		}

		// parse value
		var (
			data         string
			operatedData string
		)
		switch prop.Type {
		case v1alpha1.ModbusDevicePropertyTypeHexString:
			data = hex.EncodeToString(val)
		case v1alpha1.ModbusDevicePropertyTypeInt16:
			var valBits = visitor.Endianness.Uint16(val)
			var valFloat64 = math.Float64frombits(uint64(valBits))
			data = strconv.FormatInt(int64(valBits), 10)
			operatedData, err = doArithmeticOperations(valFloat64, visitor.OrderOfOperations)
			if err != nil {
				return "", "", errors.Wrap(err, "failed to execute arithmetic operations")
			}
		case v1alpha1.ModbusDevicePropertyTypeUint16:
			var valBits = visitor.Endianness.Uint16(val)
			var valFloat64 = math.Float64frombits(uint64(valBits))
			data = strconv.FormatUint(uint64(valBits), 10)
			operatedData, err = doArithmeticOperations(valFloat64, visitor.OrderOfOperations)
			if err != nil {
				return "", "", errors.Wrap(err, "failed to execute arithmetic operations")
			}
		}
		return data, operatedData, nil
	}

	// validate first
	switch prop.Type {
	case v1alpha1.ModbusDevicePropertyTypeInt, v1alpha1.ModbusDevicePropertyTypeUint,
		v1alpha1.ModbusDevicePropertyTypeInt32, v1alpha1.ModbusDevicePropertyTypeUint32,
		v1alpha1.ModbusDevicePropertyTypeInt64, v1alpha1.ModbusDevicePropertyTypeUint64:
		// pass
	case v1alpha1.ModbusDevicePropertyTypeFloat, v1alpha1.ModbusDevicePropertyTypeDouble:
		// pass
	default:
		return "", "", errors.Errorf("multiple 16-bits quantities %s cannot set as %s type", visitor.Register, prop.Type)
	}

	// read from multiple quantities register
	var val, err = read(visitor.Offset, visitor.Quantity)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to read from multiple quantities %s", visitor.Register)
	}
	if len(val) <= 2 {
		return "", "", errors.Errorf("failed to read from multiple 16-bits quantities %s, response bytes isn't in valid size", visitor.Register)
	}

	// parse value
	var (
		data         string
		operatedData string
	)
	switch prop.Type {
	case v1alpha1.ModbusDevicePropertyTypeInt, v1alpha1.ModbusDevicePropertyTypeInt32:
		var valBits = visitor.Endianness.Uint32(val)
		var valFloat32 = math.Float32frombits(valBits)
		data = strconv.FormatInt(int64(valBits), 10)
		operatedData, err = doArithmeticOperations(float64(valFloat32), visitor.OrderOfOperations)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to execute arithmetic operations")
		}
	case v1alpha1.ModbusDevicePropertyTypeUint, v1alpha1.ModbusDevicePropertyTypeUint32:
		var valBits = visitor.Endianness.Uint32(val)
		var valFloat32 = math.Float32frombits(valBits)
		data = strconv.FormatUint(uint64(valBits), 10)
		operatedData, err = doArithmeticOperations(float64(valFloat32), visitor.OrderOfOperations)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to execute arithmetic operations")
		}
	case v1alpha1.ModbusDevicePropertyTypeInt64:
		var valBits = visitor.Endianness.Uint64(val)
		var valFloat64 = math.Float64frombits(valBits)
		data = strconv.FormatInt(int64(valBits), 10)
		operatedData, err = doArithmeticOperations(valFloat64, visitor.OrderOfOperations)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to execute arithmetic operations")
		}
	case v1alpha1.ModbusDevicePropertyTypeUint64:
		var valBits = visitor.Endianness.Uint64(val)
		var valFloat64 = math.Float64frombits(valBits)
		data = strconv.FormatUint(valBits, 10)
		operatedData, err = doArithmeticOperations(valFloat64, visitor.OrderOfOperations)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to execute arithmetic operations")
		}
	case v1alpha1.ModbusDevicePropertyTypeFloat:
		var valBits = visitor.Endianness.Uint32(val)
		var valFloat32 = math.Float32frombits(valBits)
		data = fmt.Sprint(valFloat32)
		operatedData, err = doArithmeticOperations(float64(valFloat32), visitor.OrderOfOperations)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to execute arithmetic operations")
		}
	case v1alpha1.ModbusDevicePropertyTypeDouble:
		var valBits = visitor.Endianness.Uint64(val)
		var valFloat64 = math.Float64frombits(valBits)
		data = fmt.Sprint(valFloat64)
		operatedData, err = doArithmeticOperations(valFloat64, visitor.OrderOfOperations)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to execute arithmetic operations")
		}
	}
	return data, operatedData, nil
}

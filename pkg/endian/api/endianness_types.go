package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
)

// DevicePropertyValueEndianness defines the endian of the property value.
// +kubebuilder:validation:Enum=BigEndian;BigEndianSwap;LittleEndian;LittleEndianSwap
type DevicePropertyValueEndianness string

const (
	DevicePropertyValueEndiannessBigEndian        DevicePropertyValueEndianness = "BigEndian"
	DevicePropertyValueEndiannessBigEndianSwap    DevicePropertyValueEndianness = "BigEndianSwap"
	DevicePropertyValueEndiannessLittleEndian     DevicePropertyValueEndianness = "LittleEndian"
	DevicePropertyValueEndiannessLittleEndianSwap DevicePropertyValueEndianness = "LittleEndianSwap"
)

type StringValue interface {
	fmt.Stringer
	Float64() float64
}

type stringValue struct {
	ptr interface{}
}

func (sv stringValue) String() string {
	if reflect.ValueOf(sv.ptr).IsNil() {
		return ""
	}

	switch v := sv.ptr.(type) {
	case *int8:
		return strconv.FormatInt(int64(*v), 10)
	case *int16:
		return strconv.FormatInt(int64(*v), 10)
	case *int32:
		return strconv.FormatInt(int64(*v), 10)
	case *int64:
		return strconv.FormatInt(*v, 10)
	case *uint8:
		return strconv.FormatUint(uint64(*v), 10)
	case *uint16:
		return strconv.FormatUint(uint64(*v), 10)
	case *uint32:
		return strconv.FormatUint(uint64(*v), 10)
	case *uint64:
		return strconv.FormatUint(*v, 10)
	case *float32:
		return strconv.FormatFloat(float64(*v), 'g', -1, 32)
	case *float64:
		return strconv.FormatFloat(*v, 'g', -1, 32)
	case *bool:
		return strconv.FormatBool(*v)
	}
	return ""
}

func (sv stringValue) Float64() float64 {
	if reflect.ValueOf(sv.ptr).IsNil() {
		return 0
	}

	switch v := sv.ptr.(type) {
	case *int8:
		return float64(*v)
	case *int16:
		return float64(*v)
	case *int32:
		return float64(*v)
	case *int64:
		return float64(*v)
	case *uint8:
		return float64(*v)
	case *uint16:
		return float64(*v)
	case *uint32:
		return float64(*v)
	case *uint64:
		return float64(*v)
	case *float32:
		return float64(*v)
	case *float64:
		return *v
	case *bool:
		if *v {
			return 1
		}
		return 0
	}
	return 0
}

func (e DevicePropertyValueEndianness) ConvertInt8String(str string) ([]byte, error) {
	var val, err = strconv.ParseInt(str, 10, 8)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert the string value to int8")
	}
	return e.ConvertInt8(int8(val))
}

func (e DevicePropertyValueEndianness) ConvertInt8(i int8) ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 1))
	if err := e.Convert(buf, i); err != nil {
		return nil, errors.Wrap(err, "failed to convert int8 to bytes")
	}
	return buf.Bytes(), nil
}

func (e DevicePropertyValueEndianness) ConvertInt16String(str string) ([]byte, error) {
	var val, err = strconv.ParseInt(str, 10, 16)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert the string value to int16")
	}
	return e.ConvertInt16(int16(val))
}

func (e DevicePropertyValueEndianness) ConvertInt16(i int16) ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 2))
	if err := e.Convert(buf, i); err != nil {
		return nil, errors.Wrap(err, "failed to convert int16 to bytes")
	}
	return buf.Bytes(), nil
}

func (e DevicePropertyValueEndianness) ConvertInt32String(str string) ([]byte, error) {
	var val, err = strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert the string value to int32")
	}
	return e.ConvertInt32(int32(val))
}

func (e DevicePropertyValueEndianness) ConvertInt32(i int32) ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 4))
	if err := e.Convert(buf, i); err != nil {
		return nil, errors.Wrap(err, "failed to convert int32 to bytes")
	}
	return buf.Bytes(), nil
}

func (e DevicePropertyValueEndianness) ConvertInt64String(str string) ([]byte, error) {
	var val, err = strconv.ParseInt(str, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert the string value to int64")
	}
	return e.ConvertInt64(val)
}

func (e DevicePropertyValueEndianness) ConvertInt64(i int64) ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 8))
	if err := e.Convert(buf, i); err != nil {
		return nil, errors.Wrap(err, "failed to convert int64 to bytes")
	}
	return buf.Bytes(), nil
}

func (e DevicePropertyValueEndianness) ConvertUint8String(str string) ([]byte, error) {
	var val, err = strconv.ParseUint(str, 10, 8)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert the string value to uint8")
	}
	return e.ConvertUint8(uint8(val))
}

func (e DevicePropertyValueEndianness) ConvertUint8(i uint8) ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 1))
	if err := e.Convert(buf, i); err != nil {
		return nil, errors.Wrap(err, "failed to convert uint8 to bytes")
	}
	return buf.Bytes(), nil
}

func (e DevicePropertyValueEndianness) ConvertUint16String(str string) ([]byte, error) {
	var val, err = strconv.ParseUint(str, 10, 16)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert the string value to uint16")
	}
	return e.ConvertUint16(uint16(val))
}

func (e DevicePropertyValueEndianness) ConvertUint16(i uint16) ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 2))
	if err := e.Convert(buf, i); err != nil {
		return nil, errors.Wrap(err, "failed to convert uint16 to bytes")
	}
	return buf.Bytes(), nil
}

func (e DevicePropertyValueEndianness) ConvertUint32String(str string) ([]byte, error) {
	var val, err = strconv.ParseUint(str, 10, 32)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert the string value to uint32")
	}
	return e.ConvertUint32(uint32(val))
}

func (e DevicePropertyValueEndianness) ConvertUint32(i uint32) ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 4))
	if err := e.Convert(buf, i); err != nil {
		return nil, errors.Wrap(err, "failed to convert uint32 to bytes")
	}
	return buf.Bytes(), nil
}

func (e DevicePropertyValueEndianness) ConvertUint64String(str string) ([]byte, error) {
	var val, err = strconv.ParseUint(str, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert the string value to uint64")
	}
	return e.ConvertUint64(val)
}

func (e DevicePropertyValueEndianness) ConvertUint64(i uint64) ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 8))
	if err := e.Convert(buf, i); err != nil {
		return nil, errors.Wrap(err, "failed to convert uint64 to bytes")
	}
	return buf.Bytes(), nil
}

func (e DevicePropertyValueEndianness) ConvertFloat32String(str string) ([]byte, error) {
	var val, err = strconv.ParseFloat(str, 32)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert the value to float32/float")
	}
	return e.ConvertFloat32(float32(val))
}

func (e DevicePropertyValueEndianness) ConvertFloat32(f float32) ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 4))
	if err := e.Convert(buf, f); err != nil {
		return nil, errors.Wrap(err, "failed to convert float32 to bytes")
	}
	return buf.Bytes(), nil
}

func (e DevicePropertyValueEndianness) ConvertFloat64String(str string) ([]byte, error) {
	var val, err = strconv.ParseFloat(str, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert the value to float64/double")
	}
	return e.ConvertFloat64(val)
}

func (e DevicePropertyValueEndianness) ConvertFloat64(f float64) ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 8))
	if err := e.Convert(buf, f); err != nil {
		return nil, errors.Wrap(err, "failed to convert float64 to bytes")
	}
	return buf.Bytes(), nil
}

func (e DevicePropertyValueEndianness) ConvertBooleanString(str string) ([]byte, error) {
	var val, err = strconv.ParseBool(str)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert the value to boolean")
	}
	return e.ConvertBoolean(val)
}

func (e DevicePropertyValueEndianness) ConvertBoolean(b bool) ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 1))
	if err := e.Convert(buf, b); err != nil {
		return nil, errors.Wrap(err, "failed to convert boolean to bytes")
	}
	return buf.Bytes(), nil
}

func (e DevicePropertyValueEndianness) ConvertString(s string) ([]byte, error) {
	var buf bytes.Buffer
	if err := e.Convert(&buf, []int32(s)); err != nil {
		return nil, errors.Wrap(err, "failed to convert string to bytes")
	}
	return buf.Bytes(), nil
}

func (e DevicePropertyValueEndianness) ParseInt8(b []byte) (int8, error) {
	if len(b) < 1 {
		return 0, errors.New("failed to parse bytes to int8 as the array's length is invalid")
	}

	var i int8
	if err := e.Parse(bytes.NewReader(b), &i); err != nil {
		return 0, errors.Wrap(err, "failed to parse bytes to int8")
	}
	return i, nil
}

func (e DevicePropertyValueEndianness) ParseInt8StringValue(b []byte) (StringValue, error) {
	var i, err = e.ParseInt8(b)
	if err != nil {
		return nil, err
	}
	return stringValue{ptr: &i}, nil
}

func (e DevicePropertyValueEndianness) ParseInt16(b []byte) (int16, error) {
	if len(b) < 2 {
		return 0, errors.New("failed to parse bytes to int16 as the array's length is invalid")
	}

	var i int16
	if err := e.Parse(bytes.NewReader(b), &i); err != nil {
		return 0, errors.Wrap(err, "failed to parse bytes to int16")
	}
	return i, nil
}

func (e DevicePropertyValueEndianness) ParseInt16StringValue(b []byte) (StringValue, error) {
	var i, err = e.ParseInt16(b)
	if err != nil {
		return nil, err
	}
	return stringValue{ptr: &i}, nil
}

func (e DevicePropertyValueEndianness) ParseInt32(b []byte) (int32, error) {
	if len(b) < 4 {
		return 0, errors.New("failed to parse bytes to int32 as the array's length is invalid")
	}

	var i int32
	if err := e.Parse(bytes.NewReader(b), &i); err != nil {
		return 0, errors.Wrap(err, "failed to parse bytes to int32")
	}
	return i, nil
}

func (e DevicePropertyValueEndianness) ParseInt32StringValue(b []byte) (StringValue, error) {
	var i, err = e.ParseInt32(b)
	if err != nil {
		return nil, err
	}
	return stringValue{ptr: &i}, nil
}

func (e DevicePropertyValueEndianness) ParseInt64(b []byte) (int64, error) {
	if len(b) < 8 {
		return 0, errors.New("failed to parse bytes to int64 as the array's length is invalid")
	}

	var i int64
	if err := e.Parse(bytes.NewReader(b), &i); err != nil {
		return 0, errors.Wrap(err, "failed to parse bytes to int64")
	}
	return i, nil
}

func (e DevicePropertyValueEndianness) ParseInt64StringValue(b []byte) (StringValue, error) {
	var i, err = e.ParseInt64(b)
	if err != nil {
		return nil, err
	}
	return stringValue{ptr: &i}, nil
}

func (e DevicePropertyValueEndianness) ParseUint8(b []byte) (uint8, error) {
	if len(b) < 1 {
		return 0, errors.New("failed to parse bytes to uint8 as the array's length is invalid")
	}

	var i uint8
	if err := e.Parse(bytes.NewReader(b), &i); err != nil {
		return 0, errors.Wrap(err, "failed to parse bytes to uint8")
	}
	return i, nil
}

func (e DevicePropertyValueEndianness) ParseUint8StringValue(b []byte) (StringValue, error) {
	var i, err = e.ParseUint8(b)
	if err != nil {
		return nil, err
	}
	return stringValue{ptr: &i}, nil
}

func (e DevicePropertyValueEndianness) ParseUint16(b []byte) (uint16, error) {
	if len(b) < 2 {
		return 0, errors.New("failed to parse bytes to uint16 as the array's length is invalid")
	}

	var i uint16
	if err := e.Parse(bytes.NewReader(b), &i); err != nil {
		return 0, errors.Wrap(err, "failed to parse bytes to int16")
	}
	return i, nil
}

func (e DevicePropertyValueEndianness) ParseUint16StringValue(b []byte) (StringValue, error) {
	var i, err = e.ParseUint16(b)
	if err != nil {
		return nil, err
	}
	return stringValue{ptr: &i}, nil
}

func (e DevicePropertyValueEndianness) ParseUint32(b []byte) (uint32, error) {
	if len(b) < 4 {
		return 0, errors.New("failed to parse bytes to uint32 as the array's length is invalid")
	}

	var i uint32
	if err := e.Parse(bytes.NewReader(b), &i); err != nil {
		return 0, errors.Wrap(err, "failed to parse bytes to uint32")
	}
	return i, nil
}

func (e DevicePropertyValueEndianness) ParseUint32StringValue(b []byte) (StringValue, error) {
	var i, err = e.ParseUint32(b)
	if err != nil {
		return nil, err
	}
	return stringValue{ptr: &i}, nil
}

func (e DevicePropertyValueEndianness) ParseUint64(b []byte) (uint64, error) {
	if len(b) < 8 {
		return 0, errors.New("failed to parse bytes to uint64 as the array's length is invalid")
	}

	var i uint64
	if err := e.Parse(bytes.NewReader(b), &i); err != nil {
		return 0, errors.Wrap(err, "failed to parse bytes to uint64")
	}
	return i, nil
}

func (e DevicePropertyValueEndianness) ParseUint64StringValue(b []byte) (StringValue, error) {
	var i, err = e.ParseUint64(b)
	if err != nil {
		return nil, err
	}
	return stringValue{ptr: &i}, nil
}

func (e DevicePropertyValueEndianness) ParseFloat32(b []byte) (float32, error) {
	if len(b) < 4 {
		return 0, errors.New("failed to parse bytes to float32/float as the array's length is invalid")
	}

	var f float32
	if err := e.Parse(bytes.NewReader(b), &f); err != nil {
		return 0, errors.Wrap(err, "failed to parse bytes to float32/float")
	}
	return f, nil
}

func (e DevicePropertyValueEndianness) ParseFloat32StringValue(b []byte) (StringValue, error) {
	var f, err = e.ParseFloat32(b)
	if err != nil {
		return nil, err
	}
	return stringValue{ptr: &f}, nil
}

func (e DevicePropertyValueEndianness) ParseFloat64(b []byte) (float64, error) {
	if len(b) < 8 {
		return 0, errors.New("failed to parse bytes to float64/double as the array's length is invalid")
	}

	var f float64
	if err := e.Parse(bytes.NewReader(b), &f); err != nil {
		return 0, errors.Wrap(err, "failed to parse bytes to float64")
	}
	return f, nil
}

func (e DevicePropertyValueEndianness) ParseFloat64StringValue(b []byte) (StringValue, error) {
	var f, err = e.ParseFloat64(b)
	if err != nil {
		return nil, err
	}
	return stringValue{ptr: &f}, nil
}

func (e DevicePropertyValueEndianness) ParseBoolean(b []byte) (bool, error) {
	if len(b) < 1 {
		return false, errors.New("failed to parse bytes to boolean as the array's length is invalid")
	}

	var bl bool
	if err := e.Parse(bytes.NewReader(b), &bl); err != nil {
		return false, errors.Wrap(err, "failed to parse bytes to boolean")
	}
	return bl, nil
}

func (e DevicePropertyValueEndianness) ParseBooleanStringValue(b []byte) (StringValue, error) {
	var bl, err = e.ParseBoolean(b)
	if err != nil {
		return nil, err
	}
	return stringValue{ptr: &bl}, nil
}

func (e DevicePropertyValueEndianness) ParseString(b []byte) (string, error) {
	var s = make([]int32, len(b)>>2)
	if err := e.Parse(bytes.NewReader(b), s); err != nil {
		return "", errors.Wrap(err, "failed to parse bytes to string")
	}
	return string(s), nil
}

func (e DevicePropertyValueEndianness) Convert(w io.Writer, i interface{}) error {
	switch e {
	case DevicePropertyValueEndiannessBigEndian:
		return binary.Write(w, binary.BigEndian, i)
	case DevicePropertyValueEndiannessBigEndianSwap:
		return binary.Write(w, BigEndianSwap, i)
	case DevicePropertyValueEndiannessLittleEndian:
		return binary.Write(w, binary.LittleEndian, i)
	case DevicePropertyValueEndiannessLittleEndianSwap:
		return binary.Write(w, LittleEndianSwap, i)
	default:
		return errors.Errorf("error converting as illegal endian %s", e)
	}
}

func (e DevicePropertyValueEndianness) Parse(r io.Reader, i interface{}) error {
	switch e {
	case DevicePropertyValueEndiannessBigEndian:
		return binary.Read(r, binary.BigEndian, i)
	case DevicePropertyValueEndiannessBigEndianSwap:
		return binary.Read(r, BigEndianSwap, i)
	case DevicePropertyValueEndiannessLittleEndian:
		return binary.Read(r, binary.LittleEndian, i)
	case DevicePropertyValueEndiannessLittleEndianSwap:
		return binary.Read(r, LittleEndianSwap, i)
	default:
		return errors.Errorf("error parsing as illegal endian %s", e)
	}
}

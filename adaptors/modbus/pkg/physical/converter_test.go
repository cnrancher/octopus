package physical

import (
	"errors"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/rancher/octopus/adaptors/modbus/api/v1alpha1"
)

func TestByteArrayToString(t *testing.T) {
	type given struct {
		input      []byte
		dataType   v1alpha1.ModbusDevicePropertyType
		operations []v1alpha1.ModbusDeviceArithmeticOperation
	}
	type expect struct {
		result string
		err    error
	}
	var testCases = []struct {
		given  given
		expect expect
	}{
		{
			given: given{
				input:    []byte{0},
				dataType: "boolean",
			},
			expect: expect{
				result: "false",
				err:    nil,
			},
		},
		{
			given: given{
				input:    []byte{97},
				dataType: "string",
			},
			expect: expect{
				result: "a",
				err:    nil,
			},
		},
		{
			given: given{
				input:    nil,
				dataType: "",
			},
			expect: expect{
				result: "",
				err:    errors.New("invalid data type"),
			},
		},
		{
			given: given{
				input:    []byte{0, 0, 0, 0, 0, 0, 4, 210},
				dataType: "int",
			},
			expect: expect{
				result: "1234",
				err:    nil,
			},
		},
		{
			given: given{
				input:    []byte{64, 9, 33, 251, 84, 68, 45, 24},
				dataType: "float",
			},
			expect: expect{
				result: "3.141592653589793",
				err:    nil,
			},
		},
	}
	for i, tc := range testCases {
		var ret, err = ByteArrayToString(tc.given.input, tc.given.dataType, tc.given.operations)
		if !reflect.DeepEqual(ret, tc.expect.result) {
			t.Errorf("case %v: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
		if !reflect.DeepEqual(err, tc.expect.err) {
			t.Errorf("case %v: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

func TestStringToByteArray(t *testing.T) {
	type given struct {
		input    string
		dataType v1alpha1.ModbusDevicePropertyType
		length   int
	}
	type expect struct {
		result []byte
		err    error
	}
	var testCases = []struct {
		given  given
		expect expect
	}{
		{
			given: given{
				input:    "3",
				dataType: "int",
				length:   1,
			},
			expect: expect{
				result: []byte{3},
				err:    nil,
			},
		},
		{
			given: given{
				input:    "3",
				dataType: "int",
				length:   0,
			},
			expect: expect{
				result: nil,
				err:    errors.New("input is longer than target length"),
			},
		},
		{
			given: given{
				input:    "1234",
				dataType: "int",
				length:   8,
			},
			expect: expect{
				result: []byte{0, 0, 0, 0, 0, 0, 4, 210},
				err:    nil,
			},
		},
		{
			given: given{
				input:    "3.141592653589793",
				dataType: "float",
				length:   8,
			},
			expect: expect{
				result: []byte{64, 9, 33, 251, 84, 68, 45, 24},
				err:    nil,
			},
		},
	}

	for i, tc := range testCases {
		var ret, err = StringToByteArray(tc.given.input, tc.given.dataType, tc.given.length)
		if !reflect.DeepEqual(ret, tc.expect.result) {
			t.Errorf("case %v: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
		if !reflect.DeepEqual(err, tc.expect.err) {
			t.Errorf("case %v: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

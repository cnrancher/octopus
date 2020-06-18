package physical

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/rancher/octopus/adaptors/ble/api/v1alpha1"
)

func TestByteArrayToString(t *testing.T) {
	type given struct {
		converter v1alpha1.BluetoothDataConverter
		data      []byte
	}
	type expect struct {
		result float64
	}
	var testCases = []struct {
		given  given
		expect expect
	}{
		{
			given: given{
				data: []byte{0x03, 0x00, 0x01},
				converter: v1alpha1.BluetoothDataConverter{
					StartIndex:        1,
					EndIndex:          2,
					ShiftLeft:         1,
					ShiftRight:        0,
					OrderOfOperations: nil,
				},
			},
			expect: expect{
				result: 2,
			},
		},
		{
			given: given{
				data: []byte{0x00, 0x01, 0x02, 0x03},
				converter: v1alpha1.BluetoothDataConverter{
					StartIndex:        0,
					EndIndex:          3,
					ShiftLeft:         0,
					ShiftRight:        2,
					OrderOfOperations: nil,
				},
			},
			expect: expect{
				result: 72,
			},
		},
	}
	for i, tc := range testCases {
		var ret = ConvertReadData(tc.given.converter, tc.given.data)
		if !reflect.DeepEqual(ret, tc.expect.result) {
			t.Errorf("case %v: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

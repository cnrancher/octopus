package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBluetoothDeviceProperty_MergeAccessModes(t *testing.T) {
	type input struct {
		in *BluetoothDeviceProperty
	}
	type output struct {
		ret []BluetoothDevicePropertyAccessMode
	}
	var testCases = []struct {
		name     string
		given    input
		expected output
	}{
		{
			name: "BLANK_STRING",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						""},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "Notify",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeNotify},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeNotify},
			},
		},
		{
			name: "ReadOnce",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeReadOnce},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeReadOnce},
			},
		},
		{
			name: "ReadMany",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeReadMany},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "WriteOnce",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeWriteOnce},
			},
		},
		{
			name: "WriteMany",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeWriteMany},
			},
		},
		{
			name: "ReadOnce/ReadMany",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeReadOnce,
						BluetoothDevicePropertyAccessModeReadMany},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "WriteOnce/WriteMany",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeWriteOnce,
						BluetoothDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeWriteMany},
			},
		},
		{
			name: "Notify/WriteOnce/WriteMany",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeNotify,
						BluetoothDevicePropertyAccessModeWriteOnce,
						BluetoothDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeNotify,
					BluetoothDevicePropertyAccessModeWriteMany},
			},
		},
		{
			name: "Notify/ReadOnce/ReadMany/WriteOnce/WriteMany",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeNotify,
						BluetoothDevicePropertyAccessModeReadOnce,
						BluetoothDevicePropertyAccessModeReadMany,
						BluetoothDevicePropertyAccessModeWriteOnce,
						BluetoothDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeNotify,
					BluetoothDevicePropertyAccessModeWriteMany,
					BluetoothDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "ReadOnce/WriteOnce/WriteMany",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeReadOnce,
						BluetoothDevicePropertyAccessModeWriteOnce,
						BluetoothDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeWriteMany,
					BluetoothDevicePropertyAccessModeReadOnce},
			},
		},
		{
			name: "ReadOnce/ReadMany/WriteOnce",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeReadOnce,
						BluetoothDevicePropertyAccessModeReadMany,
						BluetoothDevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeWriteOnce,
					BluetoothDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "ReadOnce/ReadOnce/WriteOnce/ReadMany/ReadMany",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeReadOnce,
						BluetoothDevicePropertyAccessModeReadOnce,
						BluetoothDevicePropertyAccessModeWriteOnce,
						BluetoothDevicePropertyAccessModeReadMany,
						BluetoothDevicePropertyAccessModeReadMany},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeWriteOnce,
					BluetoothDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "BLANK_STRING/WriteOnce",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						"",
						BluetoothDevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeWriteOnce,
					BluetoothDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "Notify/WriteOnce",
			given: input{
				in: &BluetoothDeviceProperty{
					AccessModes: []BluetoothDevicePropertyAccessMode{
						BluetoothDevicePropertyAccessModeNotify,
						BluetoothDevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []BluetoothDevicePropertyAccessMode{
					BluetoothDevicePropertyAccessModeNotify,
					BluetoothDevicePropertyAccessModeWriteOnce},
			},
		},
	}

	for _, tc := range testCases {
		var actual output
		actual.ret = tc.given.in.MergeAccessModes()
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

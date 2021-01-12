package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModbusDeviceProperty_MergeAccessModes(t *testing.T) {
	type input struct {
		in *ModbusDeviceProperty
	}
	type output struct {
		ret []ModbusDevicePropertyAccessMode
	}
	var testCases = []struct {
		name     string
		given    input
		expected output
	}{
		{
			name: "BLANK_STRING",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						""},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "ReadOnce",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						ModbusDevicePropertyAccessModeReadOnce},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeReadOnce},
			},
		},
		{
			name: "ReadMany",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						ModbusDevicePropertyAccessModeReadMany},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "WriteOnce",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						ModbusDevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeWriteOnce},
			},
		},
		{
			name: "WriteMany",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						ModbusDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeWriteMany},
			},
		},
		{
			name: "ReadOnce/ReadMany",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						ModbusDevicePropertyAccessModeReadOnce,
						ModbusDevicePropertyAccessModeReadMany},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "WriteOnce/WriteMany",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						ModbusDevicePropertyAccessModeWriteOnce,
						ModbusDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeWriteMany},
			},
		},
		{
			name: "Notify/WriteOnce/WriteMany",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						ModbusDevicePropertyAccessModeWriteOnce,
						ModbusDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeWriteMany},
			},
		},
		{
			name: "Notify/ReadOnce/ReadMany/WriteOnce/WriteMany",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						ModbusDevicePropertyAccessModeReadOnce,
						ModbusDevicePropertyAccessModeReadMany,
						ModbusDevicePropertyAccessModeWriteOnce,
						ModbusDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeWriteMany,
					ModbusDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "ReadOnce/WriteOnce/WriteMany",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						ModbusDevicePropertyAccessModeReadOnce,
						ModbusDevicePropertyAccessModeWriteOnce,
						ModbusDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeWriteMany,
					ModbusDevicePropertyAccessModeReadOnce},
			},
		},
		{
			name: "ReadOnce/ReadMany/WriteOnce",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						ModbusDevicePropertyAccessModeReadOnce,
						ModbusDevicePropertyAccessModeReadMany,
						ModbusDevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeWriteOnce,
					ModbusDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "ReadOnce/ReadOnce/WriteOnce/ReadMany/ReadMany",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						ModbusDevicePropertyAccessModeReadOnce,
						ModbusDevicePropertyAccessModeReadOnce,
						ModbusDevicePropertyAccessModeWriteOnce,
						ModbusDevicePropertyAccessModeReadMany,
						ModbusDevicePropertyAccessModeReadMany},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeWriteOnce,
					ModbusDevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "BLANK_STRING/WriteOnce",
			given: input{
				in: &ModbusDeviceProperty{
					AccessModes: []ModbusDevicePropertyAccessMode{
						"",
						ModbusDevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []ModbusDevicePropertyAccessMode{
					ModbusDevicePropertyAccessModeWriteOnce,
					ModbusDevicePropertyAccessModeReadMany},
			},
		},
	}

	for _, tc := range testCases {
		var actual output
		actual.ret = tc.given.in.MergeAccessModes()
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

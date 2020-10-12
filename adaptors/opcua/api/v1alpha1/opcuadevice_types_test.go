package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOPCUADeviceProperty_MergeAccessModes(t *testing.T) {
	type input struct {
		in *OPCUADeviceProperty
	}
	type output struct {
		ret []OPCUADevicePropertyAccessMode
	}
	var testCases = []struct {
		name     string
		given    input
		expected output
	}{
		{
			name: "BLANK_STRING",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						""},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "Notify",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeNotify},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeNotify},
			},
		},
		{
			name: "ReadOnce",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeReadOnce},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeReadOnce},
			},
		},
		{
			name: "ReadMany",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeReadMany},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "WriteOnce",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeWriteOnce},
			},
		},
		{
			name: "WriteMany",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeWriteMany},
			},
		},
		{
			name: "ReadOnce/ReadMany",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeReadOnce,
						OPCUADevicePropertyAccessModeReadMany},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "WriteOnce/WriteMany",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeWriteOnce,
						OPCUADevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeWriteMany},
			},
		},
		{
			name: "Notify/WriteOnce/WriteMany",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeNotify,
						OPCUADevicePropertyAccessModeWriteOnce,
						OPCUADevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeNotify,
					OPCUADevicePropertyAccessModeWriteMany},
			},
		},
		{
			name: "Notify/ReadOnce/ReadMany/WriteOnce/WriteMany",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeNotify,
						OPCUADevicePropertyAccessModeReadOnce,
						OPCUADevicePropertyAccessModeReadMany,
						OPCUADevicePropertyAccessModeWriteOnce,
						OPCUADevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeNotify,
					OPCUADevicePropertyAccessModeWriteMany,
					OPCUADevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "ReadOnce/WriteOnce/WriteMany",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeReadOnce,
						OPCUADevicePropertyAccessModeWriteOnce,
						OPCUADevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeWriteMany,
					OPCUADevicePropertyAccessModeReadOnce},
			},
		},
		{
			name: "ReadOnce/ReadMany/WriteOnce",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeReadOnce,
						OPCUADevicePropertyAccessModeReadMany,
						OPCUADevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeWriteOnce,
					OPCUADevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "ReadOnce/ReadOnce/WriteOnce/ReadMany/ReadMany",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeReadOnce,
						OPCUADevicePropertyAccessModeReadOnce,
						OPCUADevicePropertyAccessModeWriteOnce,
						OPCUADevicePropertyAccessModeReadMany,
						OPCUADevicePropertyAccessModeReadMany},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeWriteOnce,
					OPCUADevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "BLANK_STRING/WriteOnce",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						"",
						OPCUADevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeWriteOnce,
					OPCUADevicePropertyAccessModeReadMany},
			},
		},
		{
			name: "Notify/WriteOnce",
			given: input{
				in: &OPCUADeviceProperty{
					AccessModes: []OPCUADevicePropertyAccessMode{
						OPCUADevicePropertyAccessModeNotify,
						OPCUADevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []OPCUADevicePropertyAccessMode{
					OPCUADevicePropertyAccessModeNotify,
					OPCUADevicePropertyAccessModeWriteOnce},
			},
		},
	}

	for _, tc := range testCases {
		var actual output
		actual.ret = tc.given.in.MergeAccessModes()
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

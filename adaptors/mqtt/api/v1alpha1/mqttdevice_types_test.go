package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMQTTDeviceProperty_MergeAccessModes(t *testing.T) {
	type input struct {
		in *MQTTDeviceProperty
	}
	type output struct {
		ret []MQTTDevicePropertyAccessMode
	}
	var testCases = []struct {
		name     string
		given    input
		expected output
	}{
		{
			name: "BLANK_STRING",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						""},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeNotify},
			},
		},
		{
			name: "Notify",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						MQTTDevicePropertyAccessModeNotify},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeNotify},
			},
		},
		{
			name: "ReadOnce",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						MQTTDevicePropertyAccessModeReadOnce},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeReadOnce},
			},
		},
		{
			name: "WriteOnce",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						MQTTDevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeWriteOnce},
			},
		},
		{
			name: "WriteMany",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						MQTTDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeWriteMany},
			},
		},
		{
			name: "ReadOnce/Notify",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						MQTTDevicePropertyAccessModeReadOnce,
						MQTTDevicePropertyAccessModeNotify},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeNotify},
			},
		},
		{
			name: "WriteOnce/WriteMany",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						MQTTDevicePropertyAccessModeWriteOnce,
						MQTTDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeWriteMany},
			},
		},
		{
			name: "Notify/WriteOnce/WriteMany",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						MQTTDevicePropertyAccessModeNotify,
						MQTTDevicePropertyAccessModeWriteOnce,
						MQTTDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeWriteMany,
					MQTTDevicePropertyAccessModeNotify},
			},
		},
		{
			name: "Notify/ReadOnce/WriteOnce/WriteMany",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						MQTTDevicePropertyAccessModeNotify,
						MQTTDevicePropertyAccessModeReadOnce,
						MQTTDevicePropertyAccessModeWriteOnce,
						MQTTDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeWriteMany,
					MQTTDevicePropertyAccessModeNotify},
			},
		},
		{
			name: "ReadOnce/WriteOnce/WriteMany",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						MQTTDevicePropertyAccessModeReadOnce,
						MQTTDevicePropertyAccessModeWriteOnce,
						MQTTDevicePropertyAccessModeWriteMany},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeWriteMany,
					MQTTDevicePropertyAccessModeReadOnce},
			},
		},
		{
			name: "ReadOnce/Notify/WriteOnce",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						MQTTDevicePropertyAccessModeReadOnce,
						MQTTDevicePropertyAccessModeNotify,
						MQTTDevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeWriteOnce,
					MQTTDevicePropertyAccessModeNotify},
			},
		},
		{
			name: "ReadOnce/ReadOnce/WriteOnce/Notify/Notify",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						MQTTDevicePropertyAccessModeReadOnce,
						MQTTDevicePropertyAccessModeReadOnce,
						MQTTDevicePropertyAccessModeWriteOnce,
						MQTTDevicePropertyAccessModeNotify,
						MQTTDevicePropertyAccessModeNotify},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeWriteOnce,
					MQTTDevicePropertyAccessModeNotify},
			},
		},
		{
			name: "BLANK_STRING/WriteOnce",
			given: input{
				in: &MQTTDeviceProperty{
					AccessModes: []MQTTDevicePropertyAccessMode{
						"",
						MQTTDevicePropertyAccessModeWriteOnce},
				},
			},
			expected: output{
				ret: []MQTTDevicePropertyAccessMode{
					MQTTDevicePropertyAccessModeWriteOnce,
					MQTTDevicePropertyAccessModeNotify},
			},
		},
	}

	for _, tc := range testCases {
		var actual output
		actual.ret = tc.given.in.MergeAccessModes()
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

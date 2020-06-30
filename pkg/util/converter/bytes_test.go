package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsafeStringToBytes(t *testing.T) {
	var testCases = []struct {
		name     string
		given    string
		expected []byte
	}{
		{
			name:     "std",
			given:    "octopus",
			expected: []byte("octopus"),
		},
	}

	for _, tc := range testCases {
		var actual = UnsafeStringToBytes(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

func TestUnsafeBytesToString(t *testing.T) {
	var testCases = []struct {
		name     string
		given    []byte
		expected string
	}{
		{
			name:     "std",
			given:    []byte("octopus"),
			expected: "octopus",
		},
	}

	for _, tc := range testCases {
		var ret = UnsafeBytesToString(tc.given)
		assert.Equal(t, tc.expected, ret, "case %q", tc.name)
	}
}

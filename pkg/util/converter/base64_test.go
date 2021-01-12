package converter

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeBase64String(t *testing.T) {
	type expected struct {
		ret []byte
		err error
	}

	var testCases = []struct {
		name     string
		given    string
		expected expected
	}{
		{
			name:  "std (with padded)",
			given: "Z29sYW5nICt+CiA=",
			expected: expected{
				ret: []byte("golang +~\n "),
			},
		},
		{
			name:  "none padded",
			given: "Z29sYW5nICt+CiA",
			expected: expected{
				ret: []byte("golang +~\n "),
			},
		},
		{
			name:  "uri friendly",
			given: "Z29sYW5nICt-CiA=",
			expected: expected{
				ret: []byte("golang +~\n "),
			},
		},
		{
			name:  "error string",
			given: "%%%",
			expected: expected{
				ret: nil,
				err: base64.CorruptInputError(0),
			},
		},
	}

	for _, tc := range testCases {
		var actual, actualErr = DecodeBase64String(tc.given)
		assert.Equal(t, tc.expected.ret, actual, "case %q", tc.name)
		assert.Equal(t, fmt.Sprint(tc.expected.err), fmt.Sprint(actualErr), "case %q", tc.name)
	}
}

func TestEncodeBase64(t *testing.T) {
	var testCases = []struct {
		name     string
		given    []byte
		expected []byte
	}{
		{
			name:     "std (with padded)",
			given:    []byte("golang +~\n "),
			expected: []byte("Z29sYW5nICt+CiA="),
		},
		{
			name:     "padded",
			given:    []byte("octopus"),
			expected: []byte("b2N0b3B1cw=="),
		},
	}

	for _, tc := range testCases {
		var actual = EncodeBase64(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

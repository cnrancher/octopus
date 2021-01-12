package converter

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeBinaryString(t *testing.T) {
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
			name:  "1 byte",
			given: "00100001",
			expected: expected{
				ret: []byte{33},
			},
		},
		{
			name:  "2 bytes",
			given: "1100110100000001",
			expected: expected{
				ret: []byte{205, 1},
			},
		},
		{
			name:  "error string",
			given: "0001",
			expected: expected{
				ret: nil,
				err: errors.New("the length of binary string value must be a multiple of eight"),
			},
		},
	}

	for _, tc := range testCases {
		var actual, actualErr = DecodeBinaryString(tc.given)
		assert.Equal(t, tc.expected.ret, actual, "case %q", tc.name)
		assert.Equal(t, fmt.Sprint(tc.expected.err), fmt.Sprint(actualErr), "case %q", tc.name)
	}
}

func TestEncodeBinaryToString(t *testing.T) {
	var testCases = []struct {
		name     string
		given    []byte
		expected string
	}{
		{
			name:     "1 byte",
			given:    []byte{7},
			expected: "00000111",
		},
		{
			name:     "2 bytes",
			given:    []byte{205, 1},
			expected: "1100110100000001",
		},
	}

	for _, tc := range testCases {
		var actual = EncodeBinaryToString(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

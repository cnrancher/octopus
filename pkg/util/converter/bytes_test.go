package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsafeStringToBytes(t *testing.T) {
	var testCases = []struct {
		given  string
		expect []byte
	}{
		{
			given:  "octopus",
			expect: []byte("octopus"),
		},
	}

	for i, tc := range testCases {
		var ret = UnsafeStringToBytes(tc.given)
		assert.Equal(t, tc.expect, ret, "case %v", i+1)
	}
}

func TestUnsafeBytesToString(t *testing.T) {
	var testCases = []struct {
		given  []byte
		expect string
	}{
		{
			given:  []byte("octopus"),
			expect: "octopus",
		},
	}

	for i, tc := range testCases {
		var ret = UnsafeBytesToString(tc.given)
		assert.Equal(t, tc.expect, ret, "case %v", i+1)
	}
}

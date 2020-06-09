package converter

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeBase64String(t *testing.T) {
	var testCases = []struct {
		given  string
		expect interface{}
	}{
		{ // std
			given:  "Z29sYW5nICt+CiA=",
			expect: []byte("golang +~\n "),
		},
		{ // none padded
			given:  "Z29sYW5nICt+CiA",
			expect: []byte("golang +~\n "),
		},
		{ // uri
			given:  "Z29sYW5nICt-CiA=",
			expect: []byte("golang +~\n "),
		},
		{
			given:  "%%%",
			expect: base64.CorruptInputError(0),
		},
	}

	for i, tc := range testCases {
		var ret, err = DecodeBase64String(tc.given)
		switch e := tc.expect.(type) {
		case []byte:
			assert.Equal(t, e, ret, "case %v", i+1)
		case error:
			assert.EqualError(t, err, e.Error(), "case %v", i+1)
		}
	}
}

func TestEncodeBase64(t *testing.T) {
	var testCases = []struct {
		given  []byte
		expect []byte
	}{
		{
			given:  []byte("golang +~\n "),
			expect: []byte("Z29sYW5nICt+CiA="),
		},
		{
			given:  []byte("octopus"),
			expect: []byte("b2N0b3B1cw=="),
		},
	}

	for i, tc := range testCases {
		var ret = EncodeBase64(tc.given)
		assert.Equal(t, tc.expect, ret, "case %v", i+1)
	}
}

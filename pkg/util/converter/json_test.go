package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalJSON(t *testing.T) {
	var testCases = []struct {
		given  interface{}
		expect []byte
	}{
		{
			given: map[string]string{
				"u": "u",
				"i": "i",
				"a": "a",
				"e": "e",
			},
			expect: []byte(`{"a":"a","e":"e","i":"i","u":"u"}`),
		},
		{
			given:  "test",
			expect: []byte(`"test"`),
		},
		{
			given:  1.456898920,
			expect: []byte(`1.456899`),
		},
		{
			given:  []byte{},
			expect: []byte(`""`),
		},
		{
			given:  nil,
			expect: []byte(`null`),
		},
	}

	for i, tc := range testCases {
		var ret, err = MarshalJSON(tc.given)
		if assert.Nil(t, err, "case %v", i+1) {
			assert.Equal(t, tc.expect, ret, "case %v", i+1)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	var testCases = []struct {
		given  []byte
		expect interface{}
	}{
		{
			given: []byte(`{"a":"a","e":"e","i":"i","u":"u"}`),
			expect: map[string]interface{}{
				"u": "u",
				"i": "i",
				"a": "a",
				"e": "e",
			},
		},
		{
			given:  []byte(`"test"`),
			expect: "test",
		},
		{
			given:  []byte(`1.456898920`),
			expect: 1.456898920,
		},
		{
			given:  []byte(`"1.456898920"`),
			expect: "1.456898920",
		},
		{
			given:  []byte(`null`),
			expect: nil,
		},
	}

	for i, tc := range testCases {
		var ret interface{}
		var err = UnmarshalJSON(tc.given, &ret)
		if assert.Nil(t, err, "case %v", i+1) {
			assert.Equal(t, tc.expect, ret, "case %v", i+1)
		}
	}
}

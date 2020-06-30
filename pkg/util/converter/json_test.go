package converter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalJSON(t *testing.T) {
	type expected struct {
		ret []byte
		err error
	}

	var testCases = []struct {
		name     string
		given    interface{}
		expected expected
	}{
		{
			name: "map",
			given: map[string]string{
				"u": "u",
				"i": "i",
				"a": "a",
				"e": "e",
			},
			expected: expected{
				ret: []byte(`{"a":"a","e":"e","i":"i","u":"u"}`),
			},
		},
		{
			name:  "string",
			given: "test",
			expected: expected{
				ret: []byte(`"test"`),
			},
		},
		{
			name:  "float number",
			given: 1.456898920,
			expected: expected{
				ret: []byte(`1.456899`),
			},
		},
		{
			name:  "empty byte array",
			given: []byte{},
			expected: expected{
				ret: []byte(`""`),
			},
		},
		{
			name:  "nil",
			given: nil,
			expected: expected{
				ret: []byte(`null`),
			},
		},
	}

	for _, tc := range testCases {
		var actual, actualErr = MarshalJSON(tc.given)
		assert.Equal(t, tc.expected.ret, actual, "case %q", tc.name)
		assert.Equal(t, tc.expected.err, actualErr, "case %q", tc.name)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	type given struct {
		data []byte
		v    interface{}
	}
	type expected struct {
		ret interface{}
		err error
	}

	var testCases = []struct {
		name     string
		given    given
		expected expected
	}{
		{
			name: "map",
			given: given{
				data: []byte(`{"a":"a","e":"e","i":"i","u":"u"}`),
			},
			expected: expected{
				ret: map[string]interface{}{
					"u": "u",
					"i": "i",
					"a": "a",
					"e": "e",
				},
			},
		},
		{
			name: "string",
			given: given{
				data: []byte(`"test"`),
			},
			expected: expected{
				ret: "test",
			},
		},
		{
			name: "float number",
			given: given{
				data: []byte(`1.456898920`),
			},
			expected: expected{
				ret: 1.456898920,
			},
		},
		{
			name: "float number string",
			given: given{
				data: []byte(`"1.456898920"`),
			},
			expected: expected{
				ret: "1.456898920",
			},
		},
		{
			name: "nil",
			given: given{
				data: []byte(`null`),
			},
			expected: expected{
				ret: nil,
			},
		},
	}

	for _, tc := range testCases {
		var actualErr = UnmarshalJSON(tc.given.data, &tc.given.v)
		assert.Equal(t, tc.expected.ret, tc.given.v, "case %q", tc.name)
		assert.Equal(t, fmt.Sprint(tc.expected.err), fmt.Sprint(actualErr), "case %q", tc.name)
	}
}

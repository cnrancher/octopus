package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringMapCopy(t *testing.T) {
	var testCases = []struct {
		name     string
		given    map[string]string
		expected map[string]string
	}{
		{
			name: "single key-pair",
			given: map[string]string{
				"k1": "v1",
			},
			expected: map[string]string{
				"k1": "v1",
			},
		},
		{
			name: "multiple key-pairs",
			given: map[string]string{
				"k2": "v2",
				"k3": "v3",
			},
			expected: map[string]string{
				"k3": "v3",
				"k2": "v2",
			},
		},
	}

	for _, tc := range testCases {
		var actual = StringMapCopy(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)

		t.Log("if adds a new key-pair to destination map")
		actual["xyz"] = "zyx"

		t.Log("should not effect the copied map")
		assert.NotEqual(t, tc.expected, actual, "case %q", tc.name)
	}
}

func TestStringMapCopyInto(t *testing.T) {
	type given struct {
		source      map[string]string
		destination map[string]string
	}

	var testCases = []struct {
		name     string
		given    given
		expected map[string]string
	}{
		{
			name: "same maps",
			given: given{
				source: map[string]string{
					"s1": "s1",
				},
				destination: map[string]string{
					"s1": "s1",
				},
			},
			expected: map[string]string{
				"s1": "s1",
			},
		},
		{
			name: "different maps",
			given: given{
				source: map[string]string{
					"s2": "s2",
				},
				destination: map[string]string{
					"d2": "d2",
				},
			},
			expected: map[string]string{
				"s2": "s2",
				"d2": "d2",
			},
		},
	}

	for _, tc := range testCases {
		var actual = StringMapCopyInto(tc.given.source, tc.given.destination)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)

		t.Log("if adds a new key-pair to destination map")
		actual["xyz"] = "zyx"

		t.Log("should change the destination map")
		assert.Equal(t, actual, tc.given.destination, "case %q", tc.name)

		t.Log("should not change the source map")
		assert.NotEqual(t, actual, tc.given.source, "case %q", tc.name)
	}
}

func TestDiffStringMap(t *testing.T) {
	type given struct {
		left  map[string]string
		right map[string]string
	}

	var testCases = []struct {
		name     string
		given    given
		expected bool
	}{
		{
			name: "disordered but same key-pairs maps",
			given: given{
				left: map[string]string{
					"a": "b",
					"c": "d",
				},
				right: map[string]string{
					"c": "d",
					"a": "b",
				},
			},
			expected: false,
		},
		{
			name: "different key-pairs maps",
			given: given{
				left: map[string]string{
					"a": "b",
					"c": "d",
				},
				right: map[string]string{
					"c": "d",
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		var actual = DiffStringMap(tc.given.left, tc.given.right)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

func TestFormatStringMap(t *testing.T) {
	type given struct {
		m        map[string]string
		splitter string
	}

	var testCases = []struct {
		name     string
		given    given
		expected string
	}{
		{
			name: "default splitter",
			given: given{
				m: map[string]string{
					"c": "d",
					"x": "z",
					"a": "b",
				},
				splitter: "",
			},
			expected: `a="b",c="d",x="z"`,
		},
		{
			name: "specify splitter to ';'",
			given: given{
				m: map[string]string{
					"c": "d",
					"x": "z",
					"a": "b",
				},
				splitter: ";",
			},
			expected: `a="b";c="d";x="z"`,
		},
	}

	for _, tc := range testCases {
		var actual = FormatStringMap(tc.given.m, tc.given.splitter)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

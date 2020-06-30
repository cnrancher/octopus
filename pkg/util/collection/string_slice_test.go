package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSliceContain(t *testing.T) {
	type given struct {
		slice  []string
		target string
	}

	var testCases = []struct {
		name     string
		given    given
		expected bool
	}{
		{
			name: "not-in slice",
			given: given{
				slice: []string{
					"Jimmy",
					"Gucci",
					"Kobe",
					"Jack",
				},
				target: "Frank",
			},
			expected: false,
		},
		{
			name: "in slice",
			given: given{
				slice: []string{
					"Jimmy",
					"Gucci",
					"Kobe",
					"Jack",
				},
				target: "Kobe",
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		var actual = StringSliceContain(tc.given.slice, tc.given.target)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

func TestStringSliceRemove(t *testing.T) {
	type given struct {
		slice  []string
		target string
	}

	var testCases = []struct {
		name     string
		given    given
		expected []string
	}{
		{
			name: "not-in slice",
			given: given{
				slice: []string{
					"Jimmy",
					"Gucci",
					"Kobe",
					"Jack",
				},
				target: "Frank",
			},
			expected: []string{
				"Jimmy",
				"Gucci",
				"Kobe",
				"Jack",
			},
		},
		{
			name: "in slice",
			given: given{
				slice: []string{
					"Jimmy",
					"Gucci",
					"Kobe",
					"Jack",
				},
				target: "Kobe",
			},
			expected: []string{
				"Jimmy",
				"Gucci",
				"Jack",
			},
		},
	}

	for _, tc := range testCases {
		var actual = StringSliceRemove(tc.given.slice, tc.given.target)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

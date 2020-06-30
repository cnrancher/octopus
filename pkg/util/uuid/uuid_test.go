package uuid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncate(t *testing.T) {
	type given struct {
		uuid            string
		remainingLength int
	}

	var testCases = []struct {
		name     string
		given    given
		expected string
	}{
		{
			name: "std id with 15 remaining",
			given: given{
				uuid:            "835aea2e-5f80-4d14-88f5-40c4bda41aa3",
				remainingLength: 15,
			},
			expected: "5a8044f40c41aa3",
		},
		{
			name: "zero id",
			given: given{
				uuid:            "00000000-0000-0000-0000-000000000000",
				remainingLength: 8,
			},
			expected: "00000000",
		},
		{
			name: "std id with 32 remaining",
			given: given{
				uuid:            "014997f5-1f12-498b-8631-d2f22920e20a",
				remainingLength: 32,
			},
			expected: "014997f51f12498b8631d2f22920e20a",
		},
		{
			name: "std id with 40 remaining",
			given: given{
				uuid:            "bb23d3dd-c36c-4b13-af8c-9ce8fb78dbb4",
				remainingLength: 40,
			},
			expected: "bb23d3ddc36c4b13af8c9ce8fb78dbb4",
		},
		{
			name: "std id with 0 remaining",
			given: given{
				uuid:            "41478d1e-c3f8-46e3-a3b5-ba251f285277",
				remainingLength: 0,
			},
			expected: "",
		},
	}

	for _, tc := range testCases {
		var actual = Truncate(tc.given.uuid, tc.given.remainingLength)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

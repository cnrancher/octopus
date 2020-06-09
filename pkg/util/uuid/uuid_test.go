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
		given  given
		expect string
	}{
		{
			given: given{
				uuid:            "835aea2e-5f80-4d14-88f5-40c4bda41aa3",
				remainingLength: 15,
			},
			expect: "5a8044f40c41aa3",
		},
		{
			given: given{
				uuid:            "00000000-0000-0000-0000-000000000000",
				remainingLength: 8,
			},
			expect: "00000000",
		},
		{
			given: given{
				uuid:            "014997f5-1f12-498b-8631-d2f22920e20a",
				remainingLength: 32,
			},
			expect: "014997f51f12498b8631d2f22920e20a",
		},
		{
			given: given{
				uuid:            "bb23d3dd-c36c-4b13-af8c-9ce8fb78dbb4",
				remainingLength: 40,
			},
			expect: "bb23d3ddc36c4b13af8c9ce8fb78dbb4",
		},
		{
			given: given{
				uuid:            "41478d1e-c3f8-46e3-a3b5-ba251f285277",
				remainingLength: 0,
			},
			expect: "",
		},
	}

	for i, tc := range testCases {
		var actualRet = Truncate(tc.given.uuid, tc.given.remainingLength)
		assert.Equal(t, tc.expect, actualRet, "case %v", i+1)
	}
}

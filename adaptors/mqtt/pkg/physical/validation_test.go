package physical

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_verifyWritableJSONPath(t *testing.T) {
	var testCases = []struct {
		name     string
		given    string
		expected error
	}{
		{
			name:     "escape blank",
			given:    "a\\ b.c",
			expected: nil,
		},
		{
			name:     "escape blank with minus",
			given:    "a\\ -b.c",
			expected: nil,
		},
		{
			name:     "std",
			given:    "a.b.c",
			expected: nil,
		},
		{
			name:     "with minus",
			given:    "a.b.-1",
			expected: errors.New("minus character not allowed in path"),
		},
		{
			name:     "with wildcard",
			given:    "child*.2",
			expected: errors.New("wildcard characters not allowed in path"),
		},
		{
			name:     "with wildcard",
			given:    "c?ildren.0",
			expected: errors.New("wildcard characters not allowed in path"),
		},
		{
			name:     "with modifier",
			given:    "@pretty:{\"sortKeys\":true}",
			expected: errors.New("modifiers not allowed in path"),
		},
		{
			name:     "with pipe",
			given:    "a.b|@reverse",
			expected: errors.New("pipe characters not allowed in path"),
		},
	}

	for _, tc := range testCases {
		var actual = verifyWritableJSONPath(tc.given)
		assert.Equal(t, fmt.Sprint(tc.expected), fmt.Sprint(actual), "case %q", tc.name)
	}
}

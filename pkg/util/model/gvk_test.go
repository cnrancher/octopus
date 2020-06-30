package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestGetCRDNameOfGroupVersionKind(t *testing.T) {
	var testCases = []struct {
		name     string
		given    schema.GroupVersionKind
		expected string
	}{
		{
			name: "k1s.test.io/v1",
			given: schema.GroupVersionKind{
				Group:   "test.io",
				Kind:    "K1",
				Version: "v1",
			},
			expected: "k1s.test.io",
		},
		{
			name: "k2s.test.io/v1",
			given: schema.GroupVersionKind{
				Group:   "test.io",
				Kind:    "K2",
				Version: "v1",
			},
			expected: "k2s.test.io",
		},
	}

	for _, tc := range testCases {
		var actual = GetCRDNameOfGroupVersionKind(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

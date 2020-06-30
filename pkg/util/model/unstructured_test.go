package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewInstanceOfTypeMeta(t *testing.T) {
	type expected struct {
		ret unstructured.Unstructured
		err error
	}

	var testCases = []struct {
		name     string
		given    metav1.TypeMeta
		expected expected
	}{
		{
			name: "k1s.test.io/v1",
			given: metav1.TypeMeta{
				Kind:       "K1",
				APIVersion: "test.io/v1",
			},
			expected: expected{
				ret: unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "K1",
						"apiVersion": "test.io/v1",
					},
				},
			},
		},
		{
			name: "k2s.test.io/v1alpha1",
			given: metav1.TypeMeta{
				Kind:       "K2",
				APIVersion: "test.io/v1alpha1",
			},
			expected: expected{
				ret: unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "K2",
						"apiVersion": "test.io/v1alpha1",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		var actual, actualErr = NewInstanceOfTypeMeta(tc.given)
		assert.Equal(t, tc.expected.ret, actual, "case %q", tc.name)
		assert.Equal(t, fmt.Sprint(tc.expected.err), fmt.Sprint(actualErr), "case %q", tc.name)
	}
}

func TestNewInstanceOfType(t *testing.T) {
	type expected struct {
		ret unstructured.Unstructured
		err error
	}

	var testCases = []struct {
		name     string
		given    metav1.Type
		expected expected
	}{
		{
			name: "k1s.test.io/v1",
			given: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "K1",
					"apiVersion": "test.io/v1",
				},
			},
			expected: expected{
				ret: unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "K1",
						"apiVersion": "test.io/v1",
					},
				},
			},
		},
		{
			name: "k2s.test.io/v1alpha1",
			given: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "K2",
					"apiVersion": "test.io/v1alpha1",
				},
			},
			expected: expected{
				ret: unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       "K2",
						"apiVersion": "test.io/v1alpha1",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		var actual, actualErr = NewInstanceOfType(tc.given)
		assert.Equal(t, tc.expected.ret, actual, "case %q", tc.name)
		assert.Equal(t, fmt.Sprint(tc.expected.err), fmt.Sprint(actualErr), "case %q", tc.name)
	}
}

package object

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestIsDeleted(t *testing.T) {
	var testCases = []struct {
		name     string
		given    metav1.Object
		expected bool
	}{
		{
			name: "alive instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test1",
				},
			},
			expected: false,
		},
		{
			name: "deleted instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "default",
					Name:              "test2",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		var actual = IsDeleted(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

func TestIsZero(t *testing.T) {
	var testCases = []struct {
		name     string
		given    metav1.Object
		expected bool
	}{
		{
			name: "existing instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test1",
				},
			},
			expected: false,
		},
		{
			name:     "zero instance",
			given:    &edgev1alpha1.DeviceLink{},
			expected: true,
		},
	}

	for _, tc := range testCases {
		var actual = IsZero(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

func TestIsActivating(t *testing.T) {
	var testCases = []struct {
		name     string
		given    metav1.Object
		expected bool
	}{
		{
			name: "existing instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test1",
				},
			},
			expected: true,
		},
		{
			name:     "zero instance",
			given:    &edgev1alpha1.DeviceLink{},
			expected: false,
		},
		{
			name: "deleted instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:         "default",
					Name:              "test2",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		var actual = IsActivating(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

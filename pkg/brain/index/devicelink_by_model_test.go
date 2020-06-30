package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestDeviceLinkByModelFunc(t *testing.T) {
	var testCases = []struct {
		name     string
		given    runtime.Object
		expected []string
	}{
		{
			name: "non-empty model",
			given: &edgev1alpha1.DeviceLink{
				Spec: edgev1alpha1.DeviceLinkSpec{
					Model: metav1.TypeMeta{
						Kind:       "K1",
						APIVersion: "test.io/v1",
					},
				},
			},
			expected: []string{"k1s.test.io"},
		},
		{
			name: "empty model",
			given: &edgev1alpha1.DeviceLink{
				Spec: edgev1alpha1.DeviceLinkSpec{
					Model: metav1.TypeMeta{},
				},
			},
			expected: nil,
		},
		{
			name:     "non-DeviceLink object",
			given:    &corev1.Node{},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		var actual = DeviceLinkByModelFunc(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

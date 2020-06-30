package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestDeviceLinkByNodeFunc(t *testing.T) {
	var testCases = []struct {
		name     string
		given    runtime.Object
		expected []string
	}{
		{
			name: "non-blank node name",
			given: &edgev1alpha1.DeviceLink{
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{
						Node: "edge-worker",
					},
				},
			},
			expected: []string{"edge-worker"},
		},
		{
			name: "blank node name",
			given: &edgev1alpha1.DeviceLink{
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{},
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
		var actual = DeviceLinkByNodeFunc(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

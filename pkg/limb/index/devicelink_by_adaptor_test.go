package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestDeviceLinkByAdaptorFuncFactory(t *testing.T) {
	var targetNode = "edge-worker"
	var nonTargetNode = "edge-worker1"

	var testCases = []struct {
		name     string
		given    runtime.Object
		expected []string
	}{
		{
			name: "non-empty adaptor but non-target node",
			given: &edgev1alpha1.DeviceLink{
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{
						Name: "adaptors.test.io/dummy",
					},
				},
				Status: edgev1alpha1.DeviceLinkStatus{
					NodeName: nonTargetNode,
				},
			},
			expected: nil,
		},
		{
			name: "non-empty adaptor",
			given: &edgev1alpha1.DeviceLink{
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{
						Name: "adaptors.test.io/dummy",
					},
				},
				Status: edgev1alpha1.DeviceLinkStatus{
					NodeName: targetNode,
				},
			},
			expected: []string{"adaptors.test.io/dummy"},
		},
		{
			name: "empty adaptor",
			given: &edgev1alpha1.DeviceLink{
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{},
				},
				Status: edgev1alpha1.DeviceLinkStatus{
					NodeName: targetNode,
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
		var actual = DeviceLinkByAdaptorFuncFactory(targetNode)(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

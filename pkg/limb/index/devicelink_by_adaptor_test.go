package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestDeviceLinkByAdaptorFunc(t *testing.T) {
	var testNode = "edge-worker"
	var testCases = []struct {
		name   string
		given  runtime.Object
		expect []string
	}{
		{
			name: "non-empty adaptor but requested another node",
			given: &edgev1alpha1.DeviceLink{
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{
						Name: "adaptors.test.io/dummy",
					},
				},
				Status: edgev1alpha1.DeviceLinkStatus{
					NodeName: testNode + "1",
				},
			},
			expect: nil,
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
					NodeName: testNode,
				},
			},
			expect: []string{"adaptors.test.io/dummy"},
		},
		{
			name: "empty adaptor",
			given: &edgev1alpha1.DeviceLink{
				Spec: edgev1alpha1.DeviceLinkSpec{
					Adaptor: edgev1alpha1.DeviceAdaptor{},
				},
				Status: edgev1alpha1.DeviceLinkStatus{
					NodeName: testNode,
				},
			},
			expect: nil,
		},
		{
			name:   "non-DeviceLink object",
			given:  &corev1.Node{},
			expect: nil,
		},
	}

	for _, tc := range testCases {
		var ret = DeviceLinkByAdaptorFuncFactory(testNode)(tc.given)
		assert.Equal(t, tc.expect, ret, "case %v", tc.name)
	}
}

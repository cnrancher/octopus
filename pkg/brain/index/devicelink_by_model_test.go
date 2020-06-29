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
		name   string
		given  runtime.Object
		expect []string
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
			expect: []string{"k1s.test.io"},
		},
		{
			name: "empty model",
			given: &edgev1alpha1.DeviceLink{
				Spec: edgev1alpha1.DeviceLinkSpec{
					Model: metav1.TypeMeta{},
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
		var ret = DeviceLinkByModelFunc(tc.given)
		assert.Equal(t, tc.expect, ret, "case %v", tc.name)
	}
}

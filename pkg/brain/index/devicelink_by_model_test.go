package index

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestDeviceLinkByModelFunc(t *testing.T) {
	var testCases = []struct {
		given  runtime.Object
		expect []string
	}{
		{
			given: &edgev1alpha1.DeviceLink{
				Spec: edgev1alpha1.DeviceLinkSpec{
					Model: metav1.TypeMeta{
						Kind:       "K1",
						APIVersion: "test.io/v1",
					},
				},
			},
			expect: nil,
		},
		{
			given: &edgev1alpha1.DeviceLink{
				Spec: edgev1alpha1.DeviceLinkSpec{
					Model: metav1.TypeMeta{
						Kind:       "K1",
						APIVersion: "test.io/v1",
					},
				},
				Status: edgev1alpha1.DeviceLinkStatus{
					Model: metav1.TypeMeta{
						Kind:       "K1",
						APIVersion: "test.io/v1",
					},
				},
			},
			expect: []string{"k1s.test.io"},
		},
		{ // non-DeviceLink object
			given:  &corev1.Node{},
			expect: nil,
		},
	}

	for i, tc := range testCases {
		var ret = DeviceLinkByModelFunc(tc.given)
		if !reflect.DeepEqual(ret, tc.expect) {
			t.Errorf("case %v: expected %s, got %s", i+1, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

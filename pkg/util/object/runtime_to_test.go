package object

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestToDeviceLinkObject(t *testing.T) {
	var testCases = []struct {
		name   string
		given  runtime.Object
		expect *edgev1alpha1.DeviceLink
	}{
		{
			name: "DeviceLink instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
				},
			},
			expect: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
				},
			},
		},
		{
			name: "non-DeviceLink instance",
			given: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			expect: nil,
		},
	}

	for _, tc := range testCases {
		var ret = ToDeviceLinkObject(tc.given)
		if !reflect.DeepEqual(ret, tc.expect) {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

func TestToNodeObject(t *testing.T) {
	var testCases = []struct {
		name   string
		given  runtime.Object
		expect *corev1.Node
	}{
		{
			name: "Node instance",
			given: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			expect: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
		},
		{
			name: "non-Node instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
			expect: nil,
		},
	}

	for _, tc := range testCases {
		var ret = ToNodeObject(tc.given)
		if !reflect.DeepEqual(ret, tc.expect) {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

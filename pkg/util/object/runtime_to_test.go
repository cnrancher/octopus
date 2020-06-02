package object

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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

func TestToCustomResourceDefinitionObject(t *testing.T) {
	var testCases = []struct {
		name   string
		given  runtime.Object
		expect *apiextensionsv1.CustomResourceDefinition
	}{
		{
			name: "CRD instance",
			given: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dummyspecialdevices.edge.cattle.io",
				},
			},
			expect: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dummyspecialdevices.edge.cattle.io",
				},
			},
		},
		{
			name: "non-CRD instance",
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
		var ret = ToCustomResourceDefinitionObject(tc.given)
		if !reflect.DeepEqual(ret, tc.expect) {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

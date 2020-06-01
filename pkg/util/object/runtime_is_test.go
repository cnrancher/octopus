package object

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestIsNodeObject(t *testing.T) {
	var testCases = []struct {
		name   string
		given  runtime.Object
		expect bool
	}{
		{
			name:   "Node instance",
			given:  &corev1.Node{},
			expect: true,
		},
		{
			name:   "non-node instance",
			given:  &edgev1alpha1.DeviceLink{},
			expect: false,
		},
	}

	for _, tc := range testCases {
		var ret = IsNodeObject(tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

func TestIsCustomResourceDefinitionObject(t *testing.T) {
	var testCases = []struct {
		name   string
		given  runtime.Object
		expect bool
	}{
		{
			name:   "CRD instance",
			given:  &apiextensionsv1.CustomResourceDefinition{},
			expect: true,
		},
		{
			name:   "non-CRD instance",
			given:  &edgev1alpha1.DeviceLink{},
			expect: false,
		},
	}

	for _, tc := range testCases {
		var ret = IsCustomResourceDefinitionObject(tc.given)
		if ret != tc.expect {
			t.Errorf("case %v: expected %s, got %s", tc.name, spew.Sprintf("%#v", tc.expect), spew.Sprintf("%#v", ret))
		}
	}
}

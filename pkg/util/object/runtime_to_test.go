package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestToDeviceLinkObject(t *testing.T) {
	var testCases = []struct {
		name     string
		given    runtime.Object
		expected *edgev1alpha1.DeviceLink
	}{
		{
			name: "DeviceLink instance",
			given: &edgev1alpha1.DeviceLink{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "test",
				},
			},
			expected: &edgev1alpha1.DeviceLink{
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
			expected: nil,
		},
	}

	for _, tc := range testCases {
		var actual = ToDeviceLinkObject(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

func TestToNodeObject(t *testing.T) {
	var testCases = []struct {
		name     string
		given    runtime.Object
		expected *corev1.Node
	}{
		{
			name: "Node instance",
			given: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			expected: &corev1.Node{
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
			expected: nil,
		},
	}

	for _, tc := range testCases {
		var actual = ToNodeObject(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

func TestToCustomResourceDefinitionObject(t *testing.T) {
	var testCases = []struct {
		name     string
		given    runtime.Object
		expected *apiextensionsv1.CustomResourceDefinition
	}{
		{
			name: "CRD instance",
			given: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dummyspecialdevices.edge.cattle.io",
				},
			},
			expected: &apiextensionsv1.CustomResourceDefinition{
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
			expected: nil,
		},
	}

	for _, tc := range testCases {
		var actual = ToCustomResourceDefinitionObject(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

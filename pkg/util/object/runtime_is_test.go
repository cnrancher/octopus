package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestIsNodeObject(t *testing.T) {
	var testCases = []struct {
		name     string
		given    runtime.Object
		expected bool
	}{
		{
			name:     "Node instance",
			given:    &corev1.Node{},
			expected: true,
		},
		{
			name:     "non-node instance",
			given:    &edgev1alpha1.DeviceLink{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		var actual = IsNodeObject(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

func TestIsCustomResourceDefinitionObject(t *testing.T) {
	var testCases = []struct {
		name     string
		given    runtime.Object
		expected bool
	}{
		{
			name:     "CRD instance",
			given:    &apiextensionsv1.CustomResourceDefinition{},
			expected: true,
		},
		{
			name:     "non-CRD instance",
			given:    &edgev1alpha1.DeviceLink{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		var actual = IsCustomResourceDefinitionObject(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

func TestIsDeviceLinkObject(t *testing.T) {
	var testCases = []struct {
		name     string
		given    runtime.Object
		expected bool
	}{
		{
			name:     "DeviceLink instance",
			given:    &edgev1alpha1.DeviceLink{},
			expected: true,
		},
		{
			name:     "non-DeviceLink instance",
			given:    &corev1.Node{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		var actual = IsDeviceLinkObject(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

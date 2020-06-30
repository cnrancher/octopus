package predicate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func TestModelChangedPredicate_Update(t *testing.T) {
	var testCases = []struct {
		name     string
		given    event.UpdateEvent
		expected bool
	}{
		{
			name: "without old object",
			given: generateUpdateEvent(
				nil,
				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
				},
			),
			expected: false,
		},
		{
			name: "non-CRD instance",
			given: generateUpdateEvent(
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dl1",
						Namespace: "default",
					},
				},
				&edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dl1",
						Namespace: "default",
					},
				},
			),
			expected: true,
		},
		{
			name: "compatible with previous versions",
			given: generateUpdateEvent(
				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
					Spec: apiextensionsv1.CustomResourceDefinitionSpec{
						Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
							{
								Name:    "v1alpha1",
								Served:  true,
								Storage: true,
							},
						},
					},
					Status: apiextensionsv1.CustomResourceDefinitionStatus{
						StoredVersions: []string{
							"v1alpha1",
						},
					},
				},
				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
					Spec: apiextensionsv1.CustomResourceDefinitionSpec{
						Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
							{
								Name:    "v1alpha1",
								Served:  true,
								Storage: false,
							},
							{
								Name:    "v1",
								Served:  true,
								Storage: true,
							},
						},
					},
					Status: apiextensionsv1.CustomResourceDefinitionStatus{
						StoredVersions: []string{
							"v1",
						},
					},
				},
			),
			expected: false,
		},
		{
			name: "doesn't compatible with previous versions",
			given: generateUpdateEvent(
				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
					Spec: apiextensionsv1.CustomResourceDefinitionSpec{
						Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
							{
								Name:    "v1alpha1",
								Served:  true,
								Storage: true,
							},
						},
					},
					Status: apiextensionsv1.CustomResourceDefinitionStatus{
						StoredVersions: []string{
							"v1alpha1",
						},
					},
				},
				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
					Spec: apiextensionsv1.CustomResourceDefinitionSpec{
						Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
							{
								Name:    "v1alpha1",
								Served:  false,
								Storage: false,
							},
							{
								Name:    "v1",
								Served:  true,
								Storage: true,
							},
						},
					},
					Status: apiextensionsv1.CustomResourceDefinitionStatus{
						StoredVersions: []string{
							"v1",
						},
					},
				},
			),
			expected: true,
		},
	}

	var predication = ModelChangedPredicate{}
	for _, tc := range testCases {
		var actual = predication.Update(tc.given)
		assert.Equal(t, tc.expected, actual, "case %q", tc.name)
	}
}

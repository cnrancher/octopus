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
		name   string
		given  event.UpdateEvent
		expect bool
	}{
		{
			name: "without MetaOld",
			given: event.UpdateEvent{
				MetaOld: nil,
				ObjectOld: &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
				},
				MetaNew: &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
				},
				ObjectNew: &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dummyspecialdevices.edge.cattle.io",
					},
				},
			},
			expect: false,
		},
		{
			name: "non-CRD instance",
			given: event.UpdateEvent{
				MetaNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dl1",
						Namespace: "default",
					},
				},
				ObjectNew: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dl1",
						Namespace: "default",
					},
				},
				MetaOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dl1",
						Namespace: "default",
					},
				},
				ObjectOld: &edgev1alpha1.DeviceLink{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dl1",
						Namespace: "default",
					},
				},
			},
			expect: true,
		},
		{
			name: "compatible with previous versions",
			given: event.UpdateEvent{
				MetaOld: &apiextensionsv1.CustomResourceDefinition{
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
				ObjectOld: &apiextensionsv1.CustomResourceDefinition{
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
				MetaNew: &apiextensionsv1.CustomResourceDefinition{
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
				ObjectNew: &apiextensionsv1.CustomResourceDefinition{
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
			},
			expect: false,
		},
		{
			name: "doesn't compatible with previous versions",
			given: event.UpdateEvent{
				MetaOld: &apiextensionsv1.CustomResourceDefinition{
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
				ObjectOld: &apiextensionsv1.CustomResourceDefinition{
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
				MetaNew: &apiextensionsv1.CustomResourceDefinition{
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
				ObjectNew: &apiextensionsv1.CustomResourceDefinition{
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
			},
			expect: true,
		},
	}

	var predication = ModelChangedPredicate{}
	for _, tc := range testCases {
		var ret = predication.Update(tc.given)
		assert.Equal(t, tc.expect, ret, "case %v", tc.name)
	}
}

// +build test

package crd

import (
	"fmt"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	modelutil "github.com/rancher/octopus/pkg/util/model"
)

func MakeOfTypeMeta(model metav1.TypeMeta) *apiextensionsv1.CustomResourceDefinition {
	var modelGVK = model.GroupVersionKind()
	var crdName = modelutil.GetCRDNameOfGroupVersionKind(model.GroupVersionKind())

	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     modelGVK.Kind,
				Singular: strings.ToLower(modelGVK.Kind),
				ListKind: fmt.Sprintf("%sList", modelGVK.Kind),
				Plural:   fmt.Sprintf("%ss", strings.ToLower(modelGVK.Kind)),
			},
			Group: modelGVK.Group,
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    modelGVK.Version,
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							XPreserveUnknownFields: pointer.BoolPtr(true),
						},
					},
				},
			},
		},
	}
}

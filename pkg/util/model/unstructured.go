package model

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func NewInstanceOfTypeMeta(genericType metav1.TypeMeta) (unstructured.Unstructured, error) {
	if genericType.Kind == "" || genericType.APIVersion == "" {
		return unstructured.Unstructured{}, errors.New("cannot create unstructured object as the generic typemeta is zero value")
	}
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       genericType.Kind,
			"apiVersion": genericType.APIVersion,
		},
	}, nil
}

func NewInstanceOfType(genericType metav1.Type) (unstructured.Unstructured, error) {
	if genericType.GetKind() == "" || genericType.GetAPIVersion() == "" {
		return unstructured.Unstructured{}, errors.New("cannot create unstructured object as the generic type is zero value")
	}
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       genericType.GetKind(),
			"apiVersion": genericType.GetAPIVersion(),
		},
	}, nil
}

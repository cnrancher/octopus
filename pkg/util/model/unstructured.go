package model

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func NewInstanceOfTypeMeta(genericType metav1.TypeMeta) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       genericType.Kind,
			"apiVersion": genericType.APIVersion,
		},
	}
}

func NewInstanceOfType(genericType metav1.Type) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       genericType.GetKind(),
			"apiVersion": genericType.GetAPIVersion(),
		},
	}
}

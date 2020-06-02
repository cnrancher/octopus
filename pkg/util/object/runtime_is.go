package object

import (
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func IsNodeObject(obj runtime.Object) bool {
	if obj == nil {
		return false
	}
	_, ok := obj.(*corev1.Node)
	return ok
}

func IsCustomResourceDefinitionObject(obj runtime.Object) bool {
	if obj == nil {
		return false
	}
	_, ok := obj.(*apiextensionsv1.CustomResourceDefinition)
	return ok
}

func IsDeviceLinkObject(obj runtime.Object) bool {
	if obj == nil {
		return false
	}
	_, ok := obj.(*edgev1alpha1.DeviceLink)
	return ok
}

package object

import (
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"

	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
)

func ToDeviceLinkObject(obj runtime.Object) *edgev1alpha1.DeviceLink {
	if obj != nil {
		if r, ok := obj.(*edgev1alpha1.DeviceLink); ok {
			return r
		}
	}
	return nil
}

func ToNodeObject(obj runtime.Object) *corev1.Node {
	if obj != nil {
		if r, ok := obj.(*corev1.Node); ok {
			return r
		}
	}
	return nil
}

func ToCustomResourceDefinitionObject(obj runtime.Object) *apiextensionsv1.CustomResourceDefinition {
	if obj != nil {
		if r, ok := obj.(*apiextensionsv1.CustomResourceDefinition); ok {
			return r
		}
	}
	return nil
}

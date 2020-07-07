package object

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func GetNamespacedName(obj metav1.Object) types.NamespacedName {
	if obj == nil {
		return types.NamespacedName{}
	}

	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

func GetControlledOwnerObjectReference(obj metav1.Object) corev1.ObjectReference {
	if obj == nil {
		return corev1.ObjectReference{}
	}

	var ownerRefs = obj.GetOwnerReferences()
	if len(ownerRefs) == 0 {
		return corev1.ObjectReference{}
	}
	var owner *metav1.OwnerReference
	for _, ownerRef := range ownerRefs {
		if ownerRef.Controller != nil && *ownerRef.Controller {
			owner = &ownerRef
			break
		}
	}
	if owner == nil {
		return corev1.ObjectReference{}
	}

	return corev1.ObjectReference{
		Namespace:  obj.GetNamespace(),
		Kind:       owner.Kind,
		APIVersion: owner.APIVersion,
		Name:       owner.Name,
		UID:        owner.UID,
	}
}

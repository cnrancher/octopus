package object

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func IsActivating(obj metav1.Object) bool {
	return !IsZero(obj) && !IsDeleted(obj)
}

func IsZero(obj metav1.Object) bool {
	if obj == nil {
		return true
	}
	return obj.GetName() == ""
}

func IsDeleted(obj metav1.Object) bool {
	if obj == nil {
		return true
	}
	return !obj.GetDeletionTimestamp().IsZero()
}

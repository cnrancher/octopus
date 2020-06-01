package node

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetReady(status *corev1.NodeStatus) metav1.ConditionStatus {
	return unify(nodeConditions(status.Conditions).get(corev1.NodeReady).Status)
}

type nodeConditions []corev1.NodeCondition

func (d nodeConditions) get(t corev1.NodeConditionType) corev1.NodeCondition {
	for _, c := range d {
		if c.Type == t {
			return c
		}
	}
	return corev1.NodeCondition{}
}

func unify(status corev1.ConditionStatus) metav1.ConditionStatus {
	return metav1.ConditionStatus(status)
}

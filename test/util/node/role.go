// +build test

package node

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func IsControlPlane(node *corev1.Node) bool {
	return labels.Set(node.GetLabels()).Has("node-role.kubernetes.io/master")
}

func IsEtcd(node *corev1.Node) bool {
	return labels.Set(node.GetLabels()).Has("node-role.kubernetes.io/etcd")
}

func IsWorker(node *corev1.Node) bool {
	return !IsControlPlane(node) || !IsEtcd(node)
}

func IsOnlyWorker(node *corev1.Node) bool {
	return !IsControlPlane(node) && !IsEtcd(node)
}

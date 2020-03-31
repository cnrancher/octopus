package framework

import (
	"github.com/onsi/ginkgo"
)

var isK3dCluster bool

func K3dIt(text string, body interface{}, timeout ...float64) bool {
	if isK3dCluster {
		return ginkgo.It(text, body, timeout...)
	}
	return ginkgo.XIt(text, body)
}

func init() {
	isK3dCluster = !IsUsingExistingCluster() && GetLocalClusterKind() == K3dCluster
}

package ginkgo

import (
	"github.com/onsi/ginkgo"
	"github.com/rancher/octopus/test/framework"
	"github.com/rancher/octopus/test/framework/cluster"
)

var isK3dCluster bool

func K3dIt(text string, body interface{}, timeout ...float64) bool {
	if isK3dCluster {
		return ginkgo.It(text, body, timeout...)
	}
	return ginkgo.XIt(text, body)
}

func init() {
	isK3dCluster = !framework.IsUsingExistingCluster() && framework.GetLocalClusterType() == cluster.LocalClusterTypeK3d
}

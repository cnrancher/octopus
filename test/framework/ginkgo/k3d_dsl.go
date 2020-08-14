// +build test

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

func K3dSpecify(text string, body interface{}, timeout ...float64) bool {
	if isK3dCluster {
		return ginkgo.Specify(text, body, timeout...)
	}
	return ginkgo.XSpecify(text, body)
}

func init() {
	isK3dCluster = !framework.IsUsingExistingCluster() && framework.GetCluster() == cluster.K3d
}

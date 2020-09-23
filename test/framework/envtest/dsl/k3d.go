// +build test

package dsl

import (
	"github.com/onsi/ginkgo"

	"github.com/rancher/octopus/test/framework/envtest"
	"github.com/rancher/octopus/test/framework/envtest/cluster"
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
	isK3dCluster = !envtest.IsUsingExistingCluster() && envtest.GetProvisionerType() == cluster.ProvisionerTypeK3d
}

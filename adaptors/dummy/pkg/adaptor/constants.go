package adaptor

import (
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	Name     = "adaptors.edge.cattle.io/dummy"
	Version  = "v1alpha1"
	Endpoint = "dummy.socket"
)

var (
	Log = zap.New(zap.UseDevMode(false)).WithName("adaptor").WithName(Name)
)

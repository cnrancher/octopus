package metrics

import (
	"github.com/rancher/octopus/pkg/metrics/limb"
)

// alias limb package
var (
	RegisterLimbMetrics    = limb.RegisterMetrics
	GetLimbMetricsRecorder = limb.GetMetricsRecorder
)

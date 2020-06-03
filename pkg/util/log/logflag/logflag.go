package logflag

import (
	flag "github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/rancher/octopus/pkg/util/log/zap"
)

type loggingT struct {
	verbosity    int
	asJSON       bool
	inProduction bool
}

var logging = loggingT{}

func AddFlags(fs *flag.FlagSet) {
	fs.IntVar(&logging.verbosity, "v", logging.verbosity, "The level of log verbosity, a higher verbosity level means a log message is less important(more details). The log verbosity is following the klog's conventions: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/logging.md.")
	fs.BoolVar(&logging.asJSON, "log-as-json", logging.asJSON, "Print brief json-style log.")
	fs.BoolVar(&logging.inProduction, "log-in-production", logging.inProduction, "Use the reasonable production logging configuration of zap.")
}

func Configure() {
	var logger = zap.NewLogger(logging.asJSON, logging.inProduction)
	ctrl.SetLogger(zap.WrapAsLogr(logging.verbosity, logger))
}

package logflag

import (
	flag "github.com/spf13/pflag"
	uberzap "go.uber.org/zap"
	uberzapcore "go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type loggingT struct {
	verbosity      int
	enableBriefLog bool
}

var logging = loggingT{verbosity: 1}

func AddFlags(fs *flag.FlagSet) {
	fs.IntVar(&logging.verbosity, "v", logging.verbosity, "The level of log verbosity: debug(0) > info(1) > warn(2) > error(3) > panic(4).")
	fs.BoolVar(&logging.enableBriefLog, "enable-brief-log", logging.enableBriefLog, "Print brief json-style log on console.")
}

func Configure() {
	level := uberzap.NewAtomicLevelAt(uberzapcore.Level(int8(logging.verbosity) - 1))

	ctrl.SetLogger(zap.New(
		zap.UseDevMode(!logging.enableBriefLog),
		zap.Level(&level),
	))
}

package log

import (
	"github.com/go-logr/logr"

	"github.com/rancher/octopus/pkg/util/log/zap"
)

var log = zap.NewNullLogger()

func SetLogger(logger logr.Logger) {
	log = logger
}

func GetLogger() logr.Logger {
	return log
}

func Info(msg string, keysAndValues ...interface{}) {
	log.Info(msg, keysAndValues...)
}

func Enabled() bool {
	return log.Enabled()
}

func Error(err error, msg string, keysAndValues ...interface{}) {
	log.Error(err, msg, keysAndValues...)
}

func V(level int) logr.InfoLogger {
	return log.V(level)
}

func WithValues(keysAndValues ...interface{}) logr.Logger {
	return log.WithValues(keysAndValues...)
}

func WithName(name string) logr.Logger {
	return log.WithName(name)
}

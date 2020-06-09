package mqtt

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-logr/logr"

	"github.com/rancher/octopus/pkg/util/log/zap"
)

type printer struct {
	logger logr.InfoLogger
}

func (l printer) Println(v ...interface{}) {
	l.logger.Info(fmt.Sprint(v...))
}

func (l printer) Printf(format string, v ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, v...))
}

// SetLogger sets a concrete logging implementation for all deferred Loggers.
func SetLogger(logger logr.Logger) {
	logger = logger.WithName("paho.mqtt.golang")

	mqtt.DEBUG = printer{logger: logger.V(6)}
	mqtt.WARN = printer{logger: logger}
	mqtt.ERROR = printer{logger: logger}
	mqtt.CRITICAL = printer{logger: logger}

	if loggerWrapper, ok := logger.(zap.LoggerWrapper); ok {
		mqtt.WARN = printer{logger: zap.WrapAsWarnInfoLogr(loggerWrapper.ToZapLogger())}
		mqtt.ERROR = printer{logger: zap.WrapAsErrorInfoLogr(loggerWrapper.ToZapLogger())}
		mqtt.CRITICAL = printer{logger: zap.WrapAsFatalInfoLogr(loggerWrapper.ToZapLogger())}
	}
}

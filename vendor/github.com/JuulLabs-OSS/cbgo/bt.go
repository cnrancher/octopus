package cbgo

import (
	"os"

	"github.com/sirupsen/logrus"
)

var btlog = &logrus.Logger{
	Out: os.Stderr,
	Formatter: &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.999",
	},
	Level: logrus.ErrorLevel,
}

// SetLog replaces the cbgo logger with a custom one.
func SetLog(log *logrus.Logger) {
	btlog = log
}

// SetLogLevel configures the cbgo logger to use the specified log level.
func SetLogLevel(level logrus.Level) {
	if btlog != nil {
		btlog.SetLevel(level)
	}
}

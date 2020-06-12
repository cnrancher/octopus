package zap

import (
	"github.com/go-logr/logr"
	"go.uber.org/zap"
)

// nullInfoLogger doesn't log any log.
type nullInfoLogger struct{}

func (nullInfoLogger) Enabled() bool                   { return false }
func (nullInfoLogger) Info(_ string, _ ...interface{}) {}

// DebugInfoLogger logs as Debug level.
type debugInfoLogger struct {
	z *zap.Logger
}

func (l debugInfoLogger) Enabled() bool { return true }
func (l debugInfoLogger) Info(msg string, keysAndValues ...interface{}) {
	l.z.Debug(msg, handleFields(keysAndValues)...)
}
func (l debugInfoLogger) ToZapLogger() *zap.Logger { return l.z }

// infoInfoLogger logs as Info level.
type infoInfoLogger struct {
	z *zap.Logger
}

func (l infoInfoLogger) Enabled() bool { return true }
func (l infoInfoLogger) Info(msg string, keysAndValues ...interface{}) {
	l.z.Info(msg, handleFields(keysAndValues)...)
}
func (l infoInfoLogger) ToZapLogger() *zap.Logger { return l.z }

// warnInfoLogger logs as Warn level.
type warnInfoLogger struct {
	z *zap.Logger
}

func (l warnInfoLogger) Enabled() bool { return true }
func (l warnInfoLogger) Info(msg string, keysAndValues ...interface{}) {
	l.z.Warn(msg, handleFields(keysAndValues)...)
}
func (l warnInfoLogger) ToZapLogger() *zap.Logger { return l.z }

// errorInfoLogger logs as Error level.
type errorInfoLogger struct {
	z *zap.Logger
}

func (l errorInfoLogger) Enabled() bool { return true }
func (l errorInfoLogger) Info(msg string, keysAndValues ...interface{}) {
	l.z.Error(msg, handleFields(keysAndValues)...)
}
func (l errorInfoLogger) ToZapLogger() *zap.Logger { return l.z }

// fatalInfoLogger logs as Fatal level.
type fatalInfoLogger struct {
	z *zap.Logger
}

func (l fatalInfoLogger) Enabled() bool { return true }
func (l fatalInfoLogger) Info(msg string, keysAndValues ...interface{}) {
	l.z.Fatal(msg, handleFields(keysAndValues)...)
}
func (l fatalInfoLogger) ToZapLogger() *zap.Logger { return l.z }

// NewNullInfoLogr returns a null logr.InfoLogger.
func NewNullInfoLogr() logr.InfoLogger {
	return nullInfoLogger{}
}

// WrapAsDebugInfoLogr wraps a Zap logger as a logr.InfoLogger to logs in Debug level.
func WrapAsDebugInfoLogr(z *zap.Logger) logr.InfoLogger {
	return debugInfoLogger{z: z}
}

// WrapAsInfoInfoLogr wraps a Zap logger as a logr.InfoLogger to logs in Info level.
func WrapAsInfoInfoLogr(z *zap.Logger) logr.InfoLogger {
	return infoInfoLogger{z: z}
}

// WrapAsWarnInfoLogr wraps a Zap logger as a logr.InfoLogger to logs in Warn level.
func WrapAsWarnInfoLogr(z *zap.Logger) logr.InfoLogger {
	return warnInfoLogger{z: z}
}

// WrapAsErrorInfoLogr wraps a Zap logger as a logr.InfoLogger to logs in Error level.
func WrapAsErrorInfoLogr(z *zap.Logger) logr.InfoLogger {
	return errorInfoLogger{z: z}
}

// WrapAsFatalInfoLogr wraps a Zap logger as a logr.InfoLogger to logs in Fatal level.
func WrapAsFatalInfoLogr(z *zap.Logger) logr.InfoLogger {
	return fatalInfoLogger{z: z}
}

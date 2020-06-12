package zap

import (
	"fmt"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
)

const maxInt = int(^uint(0) >> 1)

type nullLogger struct {
	nullInfoLogger
}

func (l nullLogger) Error(err error, msg string, keysAndValues ...interface{}) {}
func (l nullLogger) V(level int) logr.InfoLogger                               { return l }
func (l nullLogger) WithValues(keysAndValues ...interface{}) logr.Logger       { return l }
func (l nullLogger) WithName(name string) logr.Logger                          { return l }

type logger struct {
	v int
	z *zap.Logger
}

func (l *logger) Enabled() bool { return true }

func (l *logger) Info(msg string, keysAndValues ...interface{}) {
	l.z.Info(msg, handleFields(keysAndValues)...)
}

func (l *logger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.z.Error(msg, handleFields(keysAndValues, zap.Error(err))...)
}

func (l *logger) V(level int) logr.InfoLogger {
	if 0 <= level && level <= l.v {
		return WrapAsInfoInfoLogr(l.z)
	}
	return NewNullInfoLogr()
}

func (l *logger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return &logger{
		v: l.v,
		z: l.z.With(handleFields(keysAndValues)...),
	}
}

func (l *logger) WithName(name string) logr.Logger {
	return &logger{
		v: l.v,
		z: l.z.Named(name),
	}
}

func (l *logger) ToZapLogger() *zap.Logger {
	return l.z
}

func handleFields(args []interface{}, additional ...zap.Field) []zap.Field {
	var argSize = len(args)
	if argSize == 0 {
		return additional
	}

	var fields = make([]zap.Field, 0, argSize/2+len(additional))
	for i := 0; i < argSize; {
		var field zap.Field

		var arg = args[i]
		switch a := arg.(type) {
		case zap.Field:
			field = a
		case string:
			if i+1 < argSize {
				field = zap.Any(a, args[i+1])
				i++
			} else {
				field = zap.Any("#key$", a)
			}
		case error:
			field = zap.Any(fmt.Sprintf("#err%d", i+1), a)
		default:
			field = zap.Any(fmt.Sprintf("#key%d", i+1), a)
		}

		fields = append(fields, field)
		i++
	}
	return append(fields, additional...)
}

// NewNullInfoLogr returns a null logr.Logger.
func NewNullLogger() logr.Logger {
	return nullLogger{}
}

// WrapAsLogrWithVerbosity is the same as WrapAsLogr, but with verbosity.
func WrapAsLogrWithVerbosity(v int, z *zap.Logger) logr.Logger {
	return &logger{
		v: v,
		z: z,
	}
}

// WrapAsLogr wraps a Zap log as logr.Logger.
func WrapAsLogr(z *zap.Logger) logr.Logger {
	return WrapAsLogrWithVerbosity(maxInt, z)
}

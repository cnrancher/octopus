package zap

import (
	"fmt"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
)

type nullLogger struct {
}

func (nullLogger) Enabled() bool { return false }

func (nullLogger) Info(_ string, _ ...interface{}) {}

type infoLogger struct {
	z *zap.Logger
}

func (l *infoLogger) Enabled() bool { return true }

func (l *infoLogger) Info(msg string, keysAndValues ...interface{}) {
	l.z.Info(msg, handleFields(keysAndValues)...)
}

type zapLogger struct {
	v int
	z *zap.Logger
}

func (l *zapLogger) Enabled() bool { return true }

func (l *zapLogger) Info(msg string, keysAndValues ...interface{}) {
	l.z.Info(msg, handleFields(keysAndValues)...)
}

func (l *zapLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.z.Error(msg, handleFields(keysAndValues, zap.Error(err))...)
}

func (l *zapLogger) V(level int) logr.InfoLogger {
	if 0 <= level && level <= l.v {
		return &infoLogger{
			z: l.z,
		}
	}

	return nullLogger{}
}

func (l *zapLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return &zapLogger{
		v: l.v,
		z: l.z.With(handleFields(keysAndValues)...),
	}
}

func (l *zapLogger) WithName(name string) logr.Logger {
	return &zapLogger{
		v: l.v,
		z: l.z.Named(name),
	}
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

// WrapAsLogr wraps a Zap log as logr.Logger.
func WrapAsLogr(v int, z *zap.Logger) logr.Logger {
	return &zapLogger{
		v: v,
		z: z,
	}
}

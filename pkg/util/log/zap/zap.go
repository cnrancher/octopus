package zap

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerWrapper interface {
	ToZapLogger() *zap.Logger
}

// NewLogger creates a Zap logger.
func NewLogger(asJSON, inProduction bool) *zap.Logger {
	var zapLevel = zap.DebugLevel
	var zapWriteSyncer = zapcore.AddSync(os.Stderr)
	var zapOptions = []zap.Option{
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel),
		zap.ErrorOutput(zapWriteSyncer),
	}

	var zapEncoderConfig = zap.NewDevelopmentEncoderConfig()
	if inProduction {
		zapLevel = zap.InfoLevel
		zapEncoderConfig = zap.NewProductionEncoderConfig()
		zapOptions = append(zapOptions,
			zap.WrapCore(func(core zapcore.Core) zapcore.Core {
				return zapcore.NewSampler(core, time.Second, 100, 100)
			}),
		)
	}

	var zapEncoder = zapcore.NewConsoleEncoder(zapEncoderConfig)
	if asJSON {
		zapEncoder = zapcore.NewJSONEncoder(zapEncoderConfig)
	}

	return zap.New(zapcore.NewCore(zapEncoder, zapWriteSyncer, zapLevel), zapOptions...)
}

// NewDevelopmentLogger creates a Zap logger with development configuration.
func NewDevelopmentLogger() *zap.Logger {
	return NewLogger(false, false)
}

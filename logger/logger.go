package logger

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// A Logger provides fast, leveled, structured logging. All methods are safe
// for concurrent use, along with filter policy to synthesis logging data.
type Logger struct {
	*zap.Logger
	level         zapcore.Level
	prefix        string
	isDevelopment bool
	traceID       string
}

// NewLog create a new logger instance with default Zap logging production config
// and a logging scope based on the given name parameter. Custom logging option
// enables filter policy, correlationID and other configuration for logger.
// Logging is enabled at Info Level and above.
//
// For further logging function. please refer to: https://pkg.go.dev/go.uber.org/zap
func NewLog(name string, options ...Option) *Logger {
	result := &Logger{}
	for _, opt := range options {
		opt(result)
	}

	result.Logger = result.newZapLogger(name)

	return result
}

func (l *Logger) newZapLogger(name string) *zap.Logger {
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	zapConfig.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	if l.isDevelopment {
		zapConfig.Development = true
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	if l.level != 0 {
		zapConfig.Level = zap.NewAtomicLevelAt(l.level)
	}

	zapLogger, err := zapConfig.Build()
	if err != nil {
		log.Println(err)
		return nil
	}

	defer func() {
		_ = zapLogger.Sync()
	}()

	// Skip 1 level: the Echo wrapper method to get to the actual application code
	zapLogger = zapLogger.WithOptions(zap.AddCallerSkip(1))

	zapLogger = zapLogger.Named(name)
	if l.traceID != "" {
		zapLogger = zapLogger.With(zap.String(LogFieldTraceID, l.traceID))
	}

	return zapLogger
}

func (l *Logger) clone() *Logger {
	clonedLogger := *l
	return &clonedLogger
}

func (l *Logger) WithCtx(ctx context.Context) *Logger {
	traceId := ctx.Value(LogFieldTraceID)
	if traceId == nil {
		return l
	}

	return l.WithTraceID(fmt.Sprintf("%s", traceId))
}

func (l *Logger) WithTraceID(id string) *Logger {
	if id == "" {
		id = uuid.New().String()
	}

	cloned := l.clone()
	cloned.traceID = id
	cloned.Logger = cloned.With(zap.String(LogFieldTraceID, id))

	return cloned
}

func (l *Logger) WithErr(err error) *Logger {
	if err == nil {
		return l
	}

	cloned := l.clone()
	errFields := make([]zap.Field, 0)
	errFields = append(errFields, zap.String(LogFieldError, err.Error()))

	cloned.Logger = cloned.With(errFields...)

	return cloned
}

func (l *Logger) WithFields(fields map[string]any) *Logger {
	cloned := l.clone()
	logFields := make([]zap.Field, 0)

	for key, value := range fields {
		logFields = append(logFields, zap.Any(key, value))
	}

	cloned.Logger = cloned.With(logFields...)

	return cloned
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Sugar().Infof(format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Sugar().Warnf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Sugar().Errorf(format, args...)
}

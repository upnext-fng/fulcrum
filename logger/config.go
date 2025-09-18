package logger

import (
	"go.uber.org/zap/zapcore"
)

type logFieldCtxKey string

const (
	LogFieldTraceID   = "trace_id"
	LogFieldError     = "error"
	LogFieldLogHeader = "log_header"
)
const (
	LogFieldTraceIDCtxKey   logFieldCtxKey = "trace_id"
	LogFieldErrorCtxKey     logFieldCtxKey = "error"
	LogFieldLogHeaderCtxKey logFieldCtxKey = "log_header"
)

const (
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
	LogLevelPanic = "PANIC"
	LogLevelFatal = "FATAL"
)

var LogLevels = map[string]zapcore.Level{
	LogLevelDebug: zapcore.DebugLevel,
	LogLevelInfo:  zapcore.InfoLevel,
	LogLevelWarn:  zapcore.WarnLevel,
	LogLevelError: zapcore.ErrorLevel,
	LogLevelPanic: zapcore.PanicLevel,
	LogLevelFatal: zapcore.FatalLevel,
}

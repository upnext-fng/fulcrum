package logger

import (
	"go.uber.org/zap/zapcore"
)

type Option func(*Logger)

func WithDevelopment(isDevelop bool) Option {
	return func(l *Logger) {
		l.isDevelopment = isDevelop
	}
}

func WithLevel(level string) Option {
	logLevel := zapcore.InfoLevel

	if lvl, exist := LogLevels[level]; exist {
		logLevel = lvl
	}

	return func(l *Logger) {
		l.level = logLevel
	}
}

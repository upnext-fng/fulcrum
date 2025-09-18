package db

import (
	"context"
	"time"

	"github.com/upnext-fng/fulcrum/logger"
	gormLogger "gorm.io/gorm/logger"
)

// GormLogger implements gorm.io/gorm/logger.Interface using the pkg logger
type GormLogger struct {
	*logger.Logger
	config gormLogger.Config
}

// NewGormLogger creates a new GORM logger adapter
func NewGormLogger(log *logger.Logger, config gormLogger.Config) *GormLogger {
	return &GormLogger{
		Logger: log,
		config: config,
	}
}

// LogMode sets the log level and returns the logger
func (gl *GormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newLogger := gl.clone()
	newLogger.config.LogLevel = level
	return newLogger
}

// Info logs info level messages
func (gl *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if gl.config.LogLevel >= gormLogger.Info {
		gl.WithCtx(ctx).Infof(msg, data...)
	}
}

// Warn logs warn level messages
func (gl *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if gl.config.LogLevel >= gormLogger.Warn {
		gl.WithCtx(ctx).Warnf(msg, data...)
	}
}

// Error logs error level messages
func (gl *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if gl.config.LogLevel >= gormLogger.Error {
		gl.WithCtx(ctx).Errorf(msg, data...)
	}
}

// Trace logs SQL queries with timing information
func (gl *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if gl.config.LogLevel <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := map[string]any{
		"elapsed":       elapsed,
		"rows_affected": rows,
		"sql":           sql,
	}

	if err != nil {
		fields["error"] = err.Error()
	}

	// Check if the query is slow
	if gl.config.SlowThreshold != 0 && elapsed > gl.config.SlowThreshold {
		gl.WithCtx(ctx).WithFields(fields).Warn("Slow SQL query")
		return
	}

	// Log based on error and log level
	if err != nil && gl.config.LogLevel >= gormLogger.Error {
		gl.WithCtx(ctx).WithFields(fields).Error("SQL query error")
	} else if gl.config.LogLevel >= gormLogger.Info {
		gl.WithCtx(ctx).WithFields(fields).Info("SQL query executed")
	}
}

// clone creates a copy of the GormLogger
func (gl *GormLogger) clone() *GormLogger {
	return &GormLogger{
		Logger: gl.Logger,
		config: gl.config,
	}
}

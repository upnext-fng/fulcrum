package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// ZerologWrapper wraps zerolog.Logger to implement our Logger interface
type ZerologWrapper struct {
	logger zerolog.Logger
	config *Config
}

// New creates a new logger instance
func New(config *Config) (Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Configure zerolog globally
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = config.TimeFormat

	// Set global log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer
	switch config.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// Assume it's a file path
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	}

	// Configure formatter
	if config.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: config.TimeFormat,
			NoColor:    false,
		}
	}

	// Create base logger
	logger := zerolog.New(output)

	// Add timestamp
	logger = logger.With().Timestamp().Logger()

	// Add caller information
	if config.CallerEnabled {
		logger = logger.With().Caller().Logger()
	}

	// Add service information
	logger = logger.With().
		Str("service", config.ServiceName).
		Str("version", config.ServiceVersion).
		Str("environment", config.Environment).
		Logger()

	// Add custom fields
	if len(config.Fields) > 0 {
		ctx := logger.With()
		for key, value := range config.Fields {
			ctx = ctx.Interface(key, value)
		}
		logger = ctx.Logger()
	}

	// Configure sampling
	if config.SamplingEnabled {
		logger = logger.Sample(&zerolog.BasicSampler{N: uint32(config.SamplingThereafter)})
	}

	return &ZerologWrapper{
		logger: logger,
		config: config,
	}, nil
}

// NewFromZerolog creates a wrapper from existing zerolog.Logger
func NewFromZerolog(logger zerolog.Logger, config *Config) Logger {
	if config == nil {
		config = DefaultConfig()
	}

	return &ZerologWrapper{
		logger: logger,
		config: config,
	}
}

// ==========================================
// LOGGER METHODS
// ==========================================

// Debug returns a debug level event
func (l *ZerologWrapper) Debug() Event {
	return &EventWrapper{event: l.logger.Debug()}
}

// Info returns an info level event
func (l *ZerologWrapper) Info() Event {
	return &EventWrapper{event: l.logger.Info()}
}

// Warn returns a warning level event
func (l *ZerologWrapper) Warn() Event {
	return &EventWrapper{event: l.logger.Warn()}
}

// Error returns an error level event
func (l *ZerologWrapper) Error() Event {
	return &EventWrapper{event: l.logger.Error()}
}

// Fatal returns a fatal level event
func (l *ZerologWrapper) Fatal() Event {
	return &EventWrapper{event: l.logger.Fatal()}
}

// Panic returns a panic level event
func (l *ZerologWrapper) Panic() Event {
	return &EventWrapper{event: l.logger.Panic()}
}

// Level checking methods
func (l *ZerologWrapper) DebugEnabled() bool {
	return l.logger.GetLevel() <= zerolog.DebugLevel
}

func (l *ZerologWrapper) InfoEnabled() bool {
	return l.logger.GetLevel() <= zerolog.InfoLevel
}

func (l *ZerologWrapper) WarnEnabled() bool {
	return l.logger.GetLevel() <= zerolog.WarnLevel
}

func (l *ZerologWrapper) ErrorEnabled() bool {
	return l.logger.GetLevel() <= zerolog.ErrorLevel
}

// ==========================================
// CONTEXT METHODS
// ==========================================

// With returns a new context
func (l *ZerologWrapper) With() Context {
	return &ContextWrapper{context: l.logger.With()}
}

// WithContext adds context to logger
func (l *ZerologWrapper) WithContext(ctx context.Context) Logger {
	return &ZerologWrapper{
		logger: l.logger.With().Ctx(ctx).Logger(),
		config: l.config,
	}
}

// WithComponent adds component name
func (l *ZerologWrapper) WithComponent(component string) Logger {
	return &ZerologWrapper{
		logger: l.logger.With().Str("component", component).Logger(),
		config: l.config,
	}
}

// WithModule adds module name
func (l *ZerologWrapper) WithModule(module string) Logger {
	return &ZerologWrapper{
		logger: l.logger.With().Str("module", module).Logger(),
		config: l.config,
	}
}

// WithUser adds user ID
func (l *ZerologWrapper) WithUser(userID string) Logger {
	return &ZerologWrapper{
		logger: l.logger.With().Str("user_id", userID).Logger(),
		config: l.config,
	}
}

// WithSession adds session ID
func (l *ZerologWrapper) WithSession(sessionID string) Logger {
	return &ZerologWrapper{
		logger: l.logger.With().Str("session_id", sessionID).Logger(),
		config: l.config,
	}
}

// WithRequestID adds request ID
func (l *ZerologWrapper) WithRequestID(requestID string) Logger {
	return &ZerologWrapper{
		logger: l.logger.With().Str("request_id", requestID).Logger(),
		config: l.config,
	}
}

// WithCorrelationID adds correlation ID
func (l *ZerologWrapper) WithCorrelationID(correlationID string) Logger {
	return &ZerologWrapper{
		logger: l.logger.With().Str("correlation_id", correlationID).Logger(),
		config: l.config,
	}
}

// ==========================================
// SPECIAL LOGGERS
// ==========================================

// Audit returns an audit event
func (l *ZerologWrapper) Audit() Event {
	return &EventWrapper{
		event: l.logger.Info().Str("log_type", "audit"),
	}
}

// Security returns a security event
func (l *ZerologWrapper) Security() Event {
	return &EventWrapper{
		event: l.logger.Warn().Str("log_type", "security"),
	}
}

// Performance returns a performance event
func (l *ZerologWrapper) Performance() Event {
	return &EventWrapper{
		event: l.logger.Debug().Str("log_type", "performance"),
	}
}

// Health logs a health check message
func (l *ZerologWrapper) Health(message string) {
	l.logger.Info().
		Str("log_type", "health").
		Msg(message)
}

// Metric logs a metric
func (l *ZerologWrapper) Metric(name string, value interface{}) {
	l.logger.Debug().
		Str("log_type", "metric").
		Str("metric_name", name).
		Interface("metric_value", value).
		Msg("metric recorded")
}

// ==========================================
// UTILITY METHODS
// ==========================================

// SetLevel sets the logger level
func (l *ZerologWrapper) SetLevel(level string) error {
	parsedLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	l.logger = l.logger.Level(parsedLevel)
	return nil
}

// GetLevel returns the current log level
func (l *ZerologWrapper) GetLevel() string {
	return l.logger.GetLevel().String()
}

// Clone creates a copy of the logger
func (l *ZerologWrapper) Clone() Logger {
	return &ZerologWrapper{
		logger: l.logger.With().Logger(),
		config: l.config,
	}
}

// ==========================================
// EVENT WRAPPER
// ==========================================

// EventWrapper wraps zerolog.Event to implement our Event interface
type EventWrapper struct {
	event *zerolog.Event
}

// Field methods
func (e *EventWrapper) Str(key, val string) Event {
	return &EventWrapper{event: e.event.Str(key, val)}
}

func (e *EventWrapper) Strs(key string, vals []string) Event {
	return &EventWrapper{event: e.event.Strs(key, vals)}
}

func (e *EventWrapper) Int(key string, val int) Event {
	return &EventWrapper{event: e.event.Int(key, val)}
}

func (e *EventWrapper) Int64(key string, val int64) Event {
	return &EventWrapper{event: e.event.Int64(key, val)}
}

func (e *EventWrapper) Float64(key string, val float64) Event {
	return &EventWrapper{event: e.event.Float64(key, val)}
}

func (e *EventWrapper) Bool(key string, val bool) Event {
	return &EventWrapper{event: e.event.Bool(key, val)}
}

func (e *EventWrapper) Time(key string, val time.Time) Event {
	return &EventWrapper{event: e.event.Time(key, val)}
}

func (e *EventWrapper) Dur(key string, val time.Duration) Event {
	return &EventWrapper{event: e.event.Dur(key, val)}
}

func (e *EventWrapper) Interface(key string, val interface{}) Event {
	return &EventWrapper{event: e.event.Interface(key, val)}
}

func (e *EventWrapper) Any(key string, val interface{}) Event {
	return &EventWrapper{event: e.event.Any(key, val)}
}

func (e *EventWrapper) Object(key string, obj LogObjectMarshaler) Event {
	return &EventWrapper{event: e.event.Object(key, obj)}
}

// Error methods
func (e *EventWrapper) Err(err error) Event {
	return &EventWrapper{event: e.event.Err(err)}
}

func (e *EventWrapper) AnErr(key string, err error) Event {
	return &EventWrapper{event: e.event.AnErr(key, err)}
}

func (e *EventWrapper) Errs(key string, errs []error) Event {
	return &EventWrapper{event: e.event.Errs(key, errs)}
}

// Context methods
func (e *EventWrapper) Ctx(ctx context.Context) Event {
	return &EventWrapper{event: e.event.Ctx(ctx)}
}

// Authentication specific methods
func (e *EventWrapper) UserID(userID string) Event {
	return &EventWrapper{event: e.event.Str("user_id", userID)}
}

func (e *EventWrapper) Email(email string) Event {
	return &EventWrapper{event: e.event.Str("email", email)}
}

func (e *EventWrapper) SessionID(sessionID string) Event {
	return &EventWrapper{event: e.event.Str("session_id", sessionID)}
}

func (e *EventWrapper) RequestID(requestID string) Event {
	return &EventWrapper{event: e.event.Str("request_id", requestID)}
}

func (e *EventWrapper) IPAddress(ip string) Event {
	return &EventWrapper{event: e.event.Str("ip_address", ip)}
}

func (e *EventWrapper) UserAgent(userAgent string) Event {
	return &EventWrapper{event: e.event.Str("user_agent", userAgent)}
}

// Security specific methods
func (e *EventWrapper) Action(action string) Event {
	return &EventWrapper{event: e.event.Str("action", action)}
}

func (e *EventWrapper) Resource(resource string) Event {
	return &EventWrapper{event: e.event.Str("resource", resource)}
}

func (e *EventWrapper) Permission(permission string) Event {
	return &EventWrapper{event: e.event.Str("permission", permission)}
}

func (e *EventWrapper) Role(role string) Event {
	return &EventWrapper{event: e.event.Str("role", role)}
}

// HTTP specific methods
func (e *EventWrapper) HTTPMethod(method string) Event {
	return &EventWrapper{event: e.event.Str("http_method", method)}
}

func (e *EventWrapper) HTTPPath(path string) Event {
	return &EventWrapper{event: e.event.Str("http_path", path)}
}

func (e *EventWrapper) HTTPStatus(status int) Event {
	return &EventWrapper{event: e.event.Int("http_status", status)}
}

func (e *EventWrapper) HTTPDuration(duration time.Duration) Event {
	return &EventWrapper{event: e.event.Dur("http_duration", duration)}
}

// Database specific methods
func (e *EventWrapper) Query(query string) Event {
	return &EventWrapper{event: e.event.Str("db_query", query)}
}

func (e *EventWrapper) Table(table string) Event {
	return &EventWrapper{event: e.event.Str("db_table", table)}
}

func (e *EventWrapper) RowsAffected(rows int64) Event {
	return &EventWrapper{event: e.event.Int64("db_rows_affected", rows)}
}

// Send methods
func (e *EventWrapper) Msg(msg string) {
	e.event.Msg(msg)
}

func (e *EventWrapper) Msgf(format string, v ...interface{}) {
	e.event.Msgf(format, v...)
}

func (e *EventWrapper) Send() {
	e.event.Send()
}

// ==========================================
// CONTEXT WRAPPER
// ==========================================

// ContextWrapper wraps zerolog.Context to implement our Context interface
type ContextWrapper struct {
	context zerolog.Context
}

// Field methods
func (c *ContextWrapper) Str(key, val string) Context {
	return &ContextWrapper{context: c.context.Str(key, val)}
}

func (c *ContextWrapper) Strs(key string, vals []string) Context {
	return &ContextWrapper{context: c.context.Strs(key, vals)}
}

func (c *ContextWrapper) Int(key string, val int) Context {
	return &ContextWrapper{context: c.context.Int(key, val)}
}

func (c *ContextWrapper) Int64(key string, val int64) Context {
	return &ContextWrapper{context: c.context.Int64(key, val)}
}

func (c *ContextWrapper) Float64(key string, val float64) Context {
	return &ContextWrapper{context: c.context.Float64(key, val)}
}

func (c *ContextWrapper) Bool(key string, val bool) Context {
	return &ContextWrapper{context: c.context.Bool(key, val)}
}

func (c *ContextWrapper) Time(key string, val time.Time) Context {
	return &ContextWrapper{context: c.context.Time(key, val)}
}

func (c *ContextWrapper) Dur(key string, val time.Duration) Context {
	return &ContextWrapper{context: c.context.Dur(key, val)}
}

func (c *ContextWrapper) Interface(key string, val interface{}) Context {
	return &ContextWrapper{context: c.context.Interface(key, val)}
}

func (c *ContextWrapper) Any(key string, val interface{}) Context {
	return &ContextWrapper{context: c.context.Any(key, val)}
}

func (c *ContextWrapper) Object(key string, obj LogObjectMarshaler) Context {
	return &ContextWrapper{context: c.context.Object(key, obj)}
}

func (c *ContextWrapper) Ctx(ctx context.Context) Context {
	return &ContextWrapper{context: c.context.Ctx(ctx)}
}

func (c *ContextWrapper) Logger() Logger {
	return &ZerologWrapper{
		logger: c.context.Logger(),
		config: DefaultConfig(),
	}
}

// ==========================================
// HELPER FUNCTIONS
// ==========================================

// FromContext extracts logger from context
func FromContext(ctx context.Context) Logger {
	logger := zerolog.Ctx(ctx)
	return NewFromZerolog(*logger, nil)
}

// ToContext adds logger to context
func ToContext(ctx context.Context, logger Logger) context.Context {
	if wrapper, ok := logger.(*ZerologWrapper); ok {
		return wrapper.logger.WithContext(ctx)
	}
	return ctx
}

// ParseLevel parses string level to zerolog level
func ParseLevel(level string) (zerolog.Level, error) {
	return zerolog.ParseLevel(strings.ToLower(level))
}

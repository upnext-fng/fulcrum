package logger

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// Logger defines the interface for structured logging
type Logger interface {
	// Level control
	Debug() Event
	Info() Event
	Warn() Event
	Error() Event
	Fatal() Event
	Panic() Event

	// Conditional logging
	DebugEnabled() bool
	InfoEnabled() bool
	WarnEnabled() bool
	ErrorEnabled() bool

	// Context operations
	With() Context
	WithContext(ctx context.Context) Logger
	WithComponent(component string) Logger
	WithModule(module string) Logger
	WithUser(userID string) Logger
	WithSession(sessionID string) Logger
	WithRequestID(requestID string) Logger
	WithCorrelationID(correlationID string) Logger

	// Special loggers
	Audit() Event
	Security() Event
	Performance() Event

	// Health and metrics
	Health(message string)
	Metric(name string, value interface{})

	// Utility methods
	SetLevel(level string) error
	GetLevel() string
	Clone() Logger
}

// Event defines the interface for log events
type Event interface {
	// Fields
	Str(key, val string) Event
	Strs(key string, vals []string) Event
	Int(key string, val int) Event
	Int64(key string, val int64) Event
	Float64(key string, val float64) Event
	Bool(key string, val bool) Event
	Time(key string, val time.Time) Event
	Dur(key string, val time.Duration) Event

	// Objects and interfaces
	Interface(key string, val interface{}) Event
	Any(key string, val interface{}) Event
	Object(key string, obj LogObjectMarshaler) Event

	// Error handling
	Err(err error) Event
	AnErr(key string, err error) Event
	Errs(key string, errs []error) Event

	// Context and tracing
	Ctx(ctx context.Context) Event

	// Authentication specific fields
	UserID(userID string) Event
	Email(email string) Event
	SessionID(sessionID string) Event
	RequestID(requestID string) Event
	IPAddress(ip string) Event
	UserAgent(userAgent string) Event

	// Security specific fields
	Action(action string) Event
	Resource(resource string) Event
	Permission(permission string) Event
	Role(role string) Event

	// HTTP specific fields
	HTTPMethod(method string) Event
	HTTPPath(path string) Event
	HTTPStatus(status int) Event
	HTTPDuration(duration time.Duration) Event

	// Database specific fields
	Query(query string) Event
	Table(table string) Event
	RowsAffected(rows int64) Event

	// Message and send
	Msg(msg string)
	Msgf(format string, v ...interface{})
	Send()
}

// Context defines the interface for logger context
type Context interface {
	// Fields
	Str(key, val string) Context
	Strs(key string, vals []string) Context
	Int(key string, val int) Context
	Int64(key string, val int64) Context
	Float64(key string, val float64) Context
	Bool(key string, val bool) Context
	Time(key string, val time.Time) Context
	Dur(key string, val time.Duration) Context

	// Objects
	Interface(key string, val interface{}) Context
	Any(key string, val interface{}) Context
	Object(key string, obj LogObjectMarshaler) Context

	// Context operations
	Ctx(ctx context.Context) Context

	// Build logger
	Logger() Logger
}

// LogObjectMarshaler defines an interface for objects that can marshal themselves for logging
type LogObjectMarshaler interface {
	MarshalZerologObject(e *zerolog.Event)
}

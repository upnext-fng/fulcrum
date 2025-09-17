package logger

import "time"

// Config represents logger configuration
type Config struct {
	Level              string        `mapstructure:"level"`
	Format             string        `mapstructure:"format"` // "json" or "console"
	Output             string        `mapstructure:"output"` // "stdout", "stderr", or file path
	TimeFormat         string        `mapstructure:"time_format"`
	CallerEnabled      bool          `mapstructure:"caller_enabled"`
	StacktraceEnabled  bool          `mapstructure:"stacktrace_enabled"`
	SamplingEnabled    bool          `mapstructure:"sampling_enabled"`
	SamplingTick       time.Duration `mapstructure:"sampling_tick"`
	SamplingInitial    int           `mapstructure:"sampling_initial"`
	SamplingThereafter int           `mapstructure:"sampling_thereafter"`

	// Service specific fields
	ServiceName    string `mapstructure:"service_name"`
	ServiceVersion string `mapstructure:"service_version"`
	Environment    string `mapstructure:"environment"`

	// Additional fields
	Fields map[string]interface{} `mapstructure:"fields"`
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:              "info",
		Format:             "json",
		Output:             "stdout",
		TimeFormat:         time.RFC3339,
		CallerEnabled:      true,
		StacktraceEnabled:  false,
		SamplingEnabled:    false,
		SamplingTick:       time.Second,
		SamplingInitial:    100,
		SamplingThereafter: 100,
		ServiceName:        "authn-authz-service",
		ServiceVersion:     "1.0.0",
		Environment:        "development",
		Fields:             make(map[string]interface{}),
	}
}

// DevelopmentConfig returns development-friendly configuration
func DevelopmentConfig() *Config {
	config := DefaultConfig()
	config.Level = "debug"
	config.Format = "console"
	config.CallerEnabled = true
	config.StacktraceEnabled = true
	config.Environment = "development"
	return config
}

// ProductionConfig returns production-optimized configuration
func ProductionConfig() *Config {
	config := DefaultConfig()
	config.Level = "info"
	config.Format = "json"
	config.CallerEnabled = false
	config.StacktraceEnabled = false
	config.SamplingEnabled = true
	config.Environment = "production"
	return config
}

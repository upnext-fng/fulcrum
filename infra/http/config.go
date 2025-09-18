package http

import (
	"fmt"
	"time"
)

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"` // in seconds
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	RequestsPerMinute int           `mapstructure:"requests_per_minute"`
	Burst             int           `mapstructure:"burst"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
}

// TimeoutConfig represents timeout configurations
type TimeoutConfig struct {
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// Config represents HTTP server configuration
type Config struct {
	Host      string           `mapstructure:"host"`
	Port      int              `mapstructure:"port"`
	Timeouts  TimeoutConfig    `mapstructure:"timeouts"`
	CORS      *CORSConfig      `mapstructure:"cors"`
	RateLimit *RateLimitConfig `mapstructure:"rate_limit"`
}

// Address returns the full address string for the server
func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Validate validates the HTTP server configuration
func (c Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if c.Timeouts.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}

	if c.Timeouts.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}

	if c.Timeouts.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive")
	}

	// Validate CORS configuration if provided
	if c.CORS != nil {
		if len(c.CORS.AllowOrigins) == 0 {
			return fmt.Errorf("at least one CORS origin must be specified")
		}
		if len(c.CORS.AllowMethods) == 0 {
			return fmt.Errorf("at least one CORS method must be specified")
		}
	}

	// Validate rate limiting configuration if enabled
	if c.RateLimit != nil && c.RateLimit.Enabled {
		if c.RateLimit.RequestsPerMinute <= 0 {
			return fmt.Errorf("requests per minute must be positive")
		}
		if c.RateLimit.Burst <= 0 {
			return fmt.Errorf("burst rate must be positive")
		}
		if c.RateLimit.CleanupInterval <= 0 {
			return fmt.Errorf("cleanup interval must be positive")
		}
	}

	return nil
}

// IsRateLimitEnabled returns true if rate limiting is enabled
func (c Config) IsRateLimitEnabled() bool {
	return c.RateLimit != nil && c.RateLimit.Enabled
}

// IsCORSEnabled returns true if CORS is enabled
func (c Config) IsCORSEnabled() bool {
	return c.CORS != nil && len(c.CORS.AllowOrigins) > 0
}

// DefaultConfig returns a default HTTP configuration
func DefaultConfig() Config {
	return Config{
		Host: "0.0.0.0",
		Port: 8888,
		Timeouts: TimeoutConfig{
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		CORS: &CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"*"},
			ExposeHeaders:    []string{},
			AllowCredentials: false,
			MaxAge:           86400, // 24 hours
		},
		RateLimit: &RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 60,
			Burst:             10,
			CleanupInterval:   time.Minute,
		},
	}
}

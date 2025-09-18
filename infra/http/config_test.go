package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Address(t *testing.T) {
	config := Config{
		Host: "localhost",
		Port: 8080,
	}

	assert.Equal(t, "localhost:8080", config.Address())
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid config",
			config: Config{
				Host: "localhost",
				Port: 8080,
				Timeouts: TimeoutConfig{
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "Empty host",
			config: Config{
				Host: "",
				Port: 8080,
				Timeouts: TimeoutConfig{
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "host cannot be empty",
		},
		{
			name: "Invalid port",
			config: Config{
				Host: "localhost",
				Port: 0,
				Timeouts: TimeoutConfig{
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "Invalid read timeout",
			config: Config{
				Host: "localhost",
				Port: 8080,
				Timeouts: TimeoutConfig{
					ReadTimeout:  0,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "read timeout must be positive",
		},
		{
			name: "Invalid write timeout",
			config: Config{
				Host: "localhost",
				Port: 8080,
				Timeouts: TimeoutConfig{
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 0,
					IdleTimeout:  60 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "write timeout must be positive",
		},
		{
			name: "Invalid idle timeout",
			config: Config{
				Host: "localhost",
				Port: 8080,
				Timeouts: TimeoutConfig{
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  0,
				},
			},
			wantErr: true,
			errMsg:  "idle timeout must be positive",
		},
		{
			name: "TLS enabled without cert file",
			config: Config{
				Host: "localhost",
				Port: 8080,
				Timeouts: TimeoutConfig{
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "TLS certificate file is required when TLS is enabled",
		},
		{
			name: "TLS enabled without key file",
			config: Config{
				Host: "localhost",
				Port: 8080,
				Timeouts: TimeoutConfig{
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "TLS key file is required when TLS is enabled",
		},
		{
			name: "CORS without origins",
			config: Config{
				Host: "localhost",
				Port: 8080,
				Timeouts: TimeoutConfig{
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
				CORS: &CORSConfig{
					AllowOrigins: []string{},
					AllowMethods: []string{"GET"},
				},
			},
			wantErr: true,
			errMsg:  "at least one CORS origin must be specified",
		},
		{
			name: "CORS without methods",
			config: Config{
				Host: "localhost",
				Port: 8080,
				Timeouts: TimeoutConfig{
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
				CORS: &CORSConfig{
					AllowOrigins: []string{"*"},
					AllowMethods: []string{},
				},
			},
			wantErr: true,
			errMsg:  "at least one CORS method must be specified",
		},
		{
			name: "Rate limit enabled with invalid requests per minute",
			config: Config{
				Host: "localhost",
				Port: 8080,
				Timeouts: TimeoutConfig{
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
				RateLimit: &RateLimitConfig{
					Enabled:           true,
					RequestsPerMinute: 0,
					Burst:             10,
					CleanupInterval:   time.Minute,
				},
			},
			wantErr: true,
			errMsg:  "requests per minute must be positive",
		},
		{
			name: "Rate limit enabled with invalid burst",
			config: Config{
				Host: "localhost",
				Port: 8080,
				Timeouts: TimeoutConfig{
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
				RateLimit: &RateLimitConfig{
					Enabled:           true,
					RequestsPerMinute: 60,
					Burst:             0,
					CleanupInterval:   time.Minute,
				},
			},
			wantErr: true,
			errMsg:  "burst rate must be positive",
		},
		{
			name: "Rate limit enabled with invalid cleanup interval",
			config: Config{
				Host: "localhost",
				Port: 8080,
				Timeouts: TimeoutConfig{
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
				RateLimit: &RateLimitConfig{
					Enabled:           true,
					RequestsPerMinute: 60,
					Burst:             10,
					CleanupInterval:   0,
				},
			},
			wantErr: true,
			errMsg:  "cleanup interval must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_IsRateLimitEnabled(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   bool
	}{
		{
			name:   "Rate limit nil",
			config: Config{RateLimit: nil},
			want:   false,
		},
		{
			name: "Rate limit disabled",
			config: Config{
				RateLimit: &RateLimitConfig{Enabled: false},
			},
			want: false,
		},
		{
			name: "Rate limit enabled",
			config: Config{
				RateLimit: &RateLimitConfig{Enabled: true},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.config.IsRateLimitEnabled())
		})
	}
}

func TestConfig_IsCORSEnabled(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   bool
	}{
		{
			name:   "CORS nil",
			config: Config{CORS: nil},
			want:   false,
		},
		{
			name: "CORS no origins",
			config: Config{
				CORS: &CORSConfig{AllowOrigins: []string{}},
			},
			want: false,
		},
		{
			name: "CORS with origins",
			config: Config{
				CORS: &CORSConfig{AllowOrigins: []string{"*"}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.config.IsCORSEnabled())
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "0.0.0.0", config.Host)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, 30*time.Second, config.Timeouts.ReadTimeout)
	assert.Equal(t, 30*time.Second, config.Timeouts.WriteTimeout)
	assert.Equal(t, 60*time.Second, config.Timeouts.IdleTimeout)
	assert.Equal(t, []string{"*"}, config.CORS.AllowOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, config.CORS.AllowMethods)
	assert.Equal(t, []string{"*"}, config.CORS.AllowHeaders)
	assert.False(t, config.CORS.AllowCredentials)
	assert.Equal(t, 86400, config.CORS.MaxAge)
	assert.True(t, config.RateLimit.Enabled)
	assert.Equal(t, 60, config.RateLimit.RequestsPerMinute)
	assert.Equal(t, 10, config.RateLimit.Burst)
	assert.Equal(t, time.Minute, config.RateLimit.CleanupInterval)

	// Should validate without errors
	assert.NoError(t, config.Validate())
}

package jwt

import "time"

type Config struct {
	Secret   string        `mapstructure:"jwt_secret"`
	TokenTTL time.Duration `mapstructure:"token_ttl"`
}

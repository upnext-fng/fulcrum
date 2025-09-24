package security

import "time"

type Config struct {
	JWTSecret string        `mapstructure:"jwt_secret"`
	TokenTTL  time.Duration `mapstructure:"token_ttl"`
	HashCost  int           `mapstructure:"hash_cost"`
}

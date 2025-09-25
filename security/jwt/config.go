package jwt

import "time"

type Config struct {
	Secret          string        `mapstructure:"jwt_secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
	Issuer          string        `mapstructure:"jwt_issuer"`
	Audience        string        `mapstructure:"jwt_audience"`
}

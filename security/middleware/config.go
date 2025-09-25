package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/upnext-fng/fulcrum/security/jwt"
)

type Config struct {
	Skipper   func(echo.Context) bool
	JWTConfig jwt.Config `mapstructure:"jwt"`
}

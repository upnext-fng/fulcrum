package security

import (
	"github.com/upnext-fng/fulcrum/security/jwt"
	"github.com/upnext-fng/fulcrum/security/middleware"
	"github.com/upnext-fng/fulcrum/security/password"
)

type Config struct {
	JWT        jwt.Config        `mapstructure:"jwt"`
	Password   password.Config   `mapstructure:"password"`
	Middleware middleware.Config `mapstructure:"middleware"`
}

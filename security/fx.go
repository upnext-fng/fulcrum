package security

import (
	"github.com/upnext-fng/fulcrum/security/jwt"
	"github.com/upnext-fng/fulcrum/security/middleware"
	"github.com/upnext-fng/fulcrum/security/password"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewSecurityService),
	jwt.Module,
	password.Module,
	middleware.Module,
)

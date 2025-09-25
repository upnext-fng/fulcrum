package jwt

import "go.uber.org/fx"

var Module = fx.Provide(NewJWTService)

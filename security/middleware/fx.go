package middleware

import (
	"go.uber.org/fx"
)

var Module = fx.Provide(
	fx.Annotate(
		NewService,
		fx.ParamTags(`name:"jwtService"`),
	),
)

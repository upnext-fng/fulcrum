// http/fx.go
package http

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewHTTPService),
)

package password

import "go.uber.org/fx"

var Module = fx.Provide(NewService)

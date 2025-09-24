package http

import (
	"context"

	"github.com/labstack/echo/v4"
)

type HTTPService interface {
	StartServer() error
	StopServer(ctx context.Context) error
	GetEngine() *echo.Echo
	RegisterRoutes(routes RouteRegistrar)
}

type RouteRegistrar interface {
	RegisterRoutes(e *echo.Echo)
}

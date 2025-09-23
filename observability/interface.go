package observability

import (
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type ObservabilityService interface {
	Logger() *logrus.Logger
	RequestLoggerMiddleware() echo.MiddlewareFunc
	HealthEndpoint() echo.HandlerFunc
}

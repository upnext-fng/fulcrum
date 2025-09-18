package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Timeout provides the middleware for API timeout
func Timeout(second int) echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{Timeout: time.Duration(second) * time.Second})
}

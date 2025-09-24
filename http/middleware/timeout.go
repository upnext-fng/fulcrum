package middleware

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
)

func Timeout(timeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
			defer cancel()

			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func DefaultTimeout() echo.MiddlewareFunc {
	return Timeout(30 * time.Second)
}

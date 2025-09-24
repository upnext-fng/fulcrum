// http/middleware/rate_limit.go
package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func RateLimit() echo.MiddlewareFunc {
	return middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
		rate.Limit(10), // 10 requests per second
	))
}

func RateLimitWithConfig(requestsPerSecond float64) echo.MiddlewareFunc {
	return middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
		rate.Limit(requestsPerSecond),
	))
}

package middleware

import (
	"context"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/upnext-fng/fulcrum/logger"
)

// RequestLogger provides the middleware for logging the inbound traffic request
func RequestLogger(name string) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRemoteIP:  true,
		LogHost:      true,
		LogMethod:    true,
		LogURI:       true,
		LogUserAgent: true,
		LogStatus:    true,
		LogLatency:   true,
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Request().RequestURI, "swagger")
		},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			requestId := c.Request().Header.Get(echo.HeaderXRequestID)
			if requestId == "" {
				requestId = c.Response().Header().Get(echo.HeaderXRequestID)
			}

			echoLogger := logger.NewLog(name, logger.WithLevel(logger.LogLevelDebug))

			echoLogger.WithFields(map[string]any{
				"request_id": requestId,
				"remote_ip":  v.RemoteIP,
				"host":       v.Host,
				"method":     v.Method,
				"uri":        v.URI,
				"user_agent": v.UserAgent,
				"status":     v.Status,
				"latency":    v.Latency,
				"bytes_in":   c.Request().Header.Get(echo.HeaderContentLength),
				"bytes_out":  strconv.FormatInt(c.Response().Size, 10),
			}).Debug("")

			return nil
		},
	})
}

func RequestID() echo.MiddlewareFunc {
	config := middleware.DefaultRequestIDConfig
	config.Skipper = func(c echo.Context) bool {
		return strings.Contains(c.Request().RequestURI, "swagger")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			rid := req.Header.Get(config.TargetHeader)
			if rid == "" {
				rid = config.Generator()
			}

			req.Header.Set(config.TargetHeader, rid)
			res.Header().Set(config.TargetHeader, rid)

			newCtx := context.WithValue(c.Request().Context(), logger.LogFieldTraceIDCtxKey, rid)
			c.SetRequest(c.Request().WithContext(newCtx))

			return next(c)
		}
	}
}

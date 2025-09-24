package observability

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type manager struct {
	logger *logrus.Logger
	config Config
}

func NewManager(config Config) ObservabilityService {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	if strings.ToLower(config.LogFormat) == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	logger.SetOutput(os.Stdout)

	return &manager{
		logger: logger,
		config: config,
	}
}

func (m *manager) Logger() *logrus.Logger {
	return m.logger
}

func (m *manager) RequestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			req := c.Request()
			res := c.Response()

			m.logger.WithFields(logrus.Fields{
				"method":     req.Method,
				"uri":        req.RequestURI,
				"status":     res.Status,
				"latency":    time.Since(start),
				"remote_ip":  c.RealIP(),
				"user_agent": req.UserAgent(),
			}).Info("HTTP Request")

			return err
		}
	}
}

func (m *manager) HealthEndpoint() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "microservice",
		})
	}
}

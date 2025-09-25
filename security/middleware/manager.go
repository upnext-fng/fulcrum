package middleware

import (
	"net/http"
	"strings"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/upnext-fng/fulcrum/security/jwt"
)

type manager struct {
	config     Config
	jwtService jwt.Service
}

func NewManager(config Config, jwtService jwt.Service) Service {
	return &manager{
		config:     config,
		jwtService: jwtService,
	}
}

func (m *manager) JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}

			tokenString := m.ExtractToken(c)
			if tokenString == "" {
				return m.createUnauthorizedError("missing authorization token")
			}

			token, err := m.jwtService.ValidateToken(tokenString)
			if err != nil {
				return m.createUnauthorizedError("invalid or expired token")
			}

			if !token.Valid {
				return m.createUnauthorizedError("invalid token")
			}

			claims, err := m.jwtService.ParseClaims(token)
			if err != nil {
				return m.createUnauthorizedError("invalid token claims")
			}

			c.Set("user_id", claims.UserID)
			c.Set("token", tokenString)
			c.Set("claims", claims)

			return next(c)
		}
	}
}

func (m *manager) AuthMiddleware() echo.MiddlewareFunc {
	return m.JWTMiddleware()
}

func (m *manager) CORSMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusOK)
			}

			return next(c)
		}
	}
}

func (m *manager) RateLimitMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}

func (m *manager) ExtractToken(c echo.Context) string {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

func (m *manager) ValidateToken(tokenString string) (*jwtlib.Token, error) {
	return m.jwtService.ValidateToken(tokenString)
}

func (m *manager) createUnauthorizedError(message string) error {
	return echo.NewHTTPError(http.StatusUnauthorized, message)
}

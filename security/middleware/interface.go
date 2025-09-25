package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type Service interface {
	JWTMiddleware() echo.MiddlewareFunc
	AuthMiddleware() echo.MiddlewareFunc
	CORSMiddleware() echo.MiddlewareFunc
	RateLimitMiddleware() echo.MiddlewareFunc
	ExtractToken(echo.Context) string
	ValidateToken(tokenString string) (*jwt.Token, error)
}

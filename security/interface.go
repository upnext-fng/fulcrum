package security

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type SecurityService interface {
	// JWT operations
	GenerateToken(userID string, claims map[string]interface{}) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)

	// Password operations
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error

	// Middleware
	JWTMiddleware() echo.MiddlewareFunc
}

package security

import (
	"github.com/labstack/echo/v4"
)

type SecurityService interface {
	// JWT operations - Strongly-typed methods
	GenerateToken(request TokenRequest) (TokenResponse, error)
	ValidateToken(request ValidationRequest) (ValidationResponse, error)
	RefreshAccessToken(refreshToken string) (TokenResponse, error)

	// Password operations
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
	ValidatePassword(password string) error

	// Middleware
	JWTMiddleware() echo.MiddlewareFunc
	AuthMiddleware() echo.MiddlewareFunc
	CORSMiddleware() echo.MiddlewareFunc
	RateLimitMiddleware() echo.MiddlewareFunc
}

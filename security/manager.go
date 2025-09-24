package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type manager struct {
	config Config
}

func NewManager(config Config) SecurityService {
	return &manager{
		config: config,
	}
}

func (m *manager) GenerateToken(userID string, claims map[string]interface{}) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	tokenClaims := token.Claims.(jwt.MapClaims)
	tokenClaims["user_id"] = userID
	tokenClaims["exp"] = time.Now().Add(m.config.TokenTTL).Unix()
	tokenClaims["iat"] = time.Now().Unix()

	// Add custom claims
	for key, value := range claims {
		tokenClaims[key] = value
	}

	return token.SignedString([]byte(m.config.JWTSecret))
}

func (m *manager) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.JWTSecret), nil
	})
}

func (m *manager) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), m.config.HashCost)
	return string(bytes), err
}

func (m *manager) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (m *manager) JWTMiddleware() echo.MiddlewareFunc {
	//return middleware.JWTWithConfig(middleware.JWTConfig{
	//	SigningKey: []byte(m.config.JWTSecret),
	//	ErrorHandler: func(c echo.Context, err error) error {
	//		return echo.NewHTTPError(401, "invalid or expired token")
	//	},
	//})
	return nil
}

package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type manager struct {
	config Config
}

func NewManager(config Config) Service {
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

	for key, value := range claims {
		tokenClaims[key] = value
	}

	return token.SignedString([]byte(m.config.Secret))
}

func (m *manager) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.Secret), nil
	})
}

func (m *manager) ParseClaims(token *jwt.Token) (*Claims, error) {
	if token == nil {
		return nil, fmt.Errorf("token is nil")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims type")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_id not found in claims")
	}

	customClaims := make(map[string]interface{})
	for key, value := range claims {
		if key != "user_id" && key != "exp" && key != "iat" {
			customClaims[key] = value
		}
	}

	return &Claims{
		UserID: userID,
		Custom: customClaims,
	}, nil
}
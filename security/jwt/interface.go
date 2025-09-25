package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	GenerateToken(userID string, claims map[string]interface{}) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	ParseClaims(token *jwt.Token) (*Claims, error)
}
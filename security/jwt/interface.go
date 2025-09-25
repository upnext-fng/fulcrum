package jwt

import "github.com/golang-jwt/jwt/v5"

// ClaimsParser Component interfaces
type ClaimsParser interface {
	ParseClaims(token *jwt.Token) (*Claims, error)
	ExtractClaims(claims jwt.MapClaims) (*Claims, error)
}

type TokenGenerator interface {
	GenerateAccessToken(claims Claims, opts ...TokenOption) (*SignedToken, error)
	GenerateRefreshToken(claims Claims, opts ...RefreshOption) (*SignedToken, error)
	GenerateTokenPair(claims Claims, opts ...TokenPairOption) (*TokenPair, error)
}

type TokenValidator interface {
	ValidateToken(tokenString string, opts ...ValidationOption) (*ValidatedToken, error)
	ValidateClaims(claims *Claims, config *ValidationConfig) error
}

type TokenRefresher interface {
	RefreshAccessToken(refreshToken string, opts ...RefreshOption) (*SignedToken, error)
}

// Service Main service interface
type Service interface {
	TokenGenerator
	TokenValidator
	TokenRefresher
}

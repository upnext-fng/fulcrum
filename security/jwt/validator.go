package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type tokenValidator struct {
	config       Config
	claimsParser ClaimsParser
}

func TokenValidatorProvider(config Config, parser ClaimsParser) TokenValidator {
	return NewTokenValidator(config, parser)
}

func NewTokenValidator(config Config, parser ClaimsParser) TokenValidator {
	return &tokenValidator{
		config:       config,
		claimsParser: parser,
	}
}

func (v *tokenValidator) ValidateToken(tokenString string, opts ...ValidationOption) (*ValidatedToken, error) {
	config := &ValidationConfig{}

	// Apply options
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return []byte(v.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Parse claims
	parsedClaims, err := v.claimsParser.ParseClaims(token)
	if err != nil {
		return nil, err
	}

	// Validate claims if not skipped
	if !config.SkipValidation {
		if err := v.ValidateClaims(parsedClaims, config); err != nil {
			return nil, err
		}
	}

	var expiresAt time.Time
	if parsedClaims.ExpiresAt > 0 {
		expiresAt = time.Unix(parsedClaims.ExpiresAt, 0)
	}

	return &ValidatedToken{
		Token:     token,
		Claims:    parsedClaims,
		IsValid:   true,
		ExpiresAt: expiresAt,
	}, nil
}

func (v *tokenValidator) ValidateClaims(claims *Claims, config *ValidationConfig) error {
	if claims == nil {
		return ErrInvalidClaims
	}

	// Check expiration
	if !config.SkipExpiration && claims.ExpiresAt > 0 {
		if time.Now().Unix() > claims.ExpiresAt {
			return ErrTokenExpired
		}
	}

	// Check not before
	if claims.NotBefore > 0 && time.Now().Unix() < claims.NotBefore {
		return ErrInvalidToken
	}

	// Check issuer
	if config.RequiredIssuer != "" && claims.Issuer != config.RequiredIssuer {
		return ErrInvalidToken
	}

	// Check audience
	if config.RequiredAudience != "" && claims.Audience != config.RequiredAudience {
		return ErrInvalidToken
	}

	// Check scopes
	if len(config.RequiredScopes) > 0 {
		if !v.hasRequiredScopes(claims.Scopes, config.RequiredScopes) {
			return ErrInvalidToken
		}
	}

	return nil
}

func (v *tokenValidator) hasRequiredScopes(tokenScopes, requiredScopes []string) bool {
	scopeMap := make(map[string]bool)
	for _, scope := range tokenScopes {
		scopeMap[scope] = true
	}

	for _, required := range requiredScopes {
		if !scopeMap[required] {
			return false
		}
	}

	return true
}

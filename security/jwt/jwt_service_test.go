package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_InterfaceImplementation(t *testing.T) {
	config := Config{
		Secret:          "test-secret-key-123",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour * 24,
		Issuer:          "test-issuer",
		Audience:        "test-audience",
	}

	service := NewJWTService(config)

	// Verify that the service implements all required interfaces
	assert.Implements(t, (*Service)(nil), service)
	assert.Implements(t, (*TokenGenerator)(nil), service)
	assert.Implements(t, (*TokenValidator)(nil), service)
	assert.Implements(t, (*TokenRefresher)(nil), service)
}

func TestJWTService_GenerateAccessToken(t *testing.T) {
	config := Config{
		Secret:         "test-secret-key-123",
		AccessTokenTTL: time.Hour,
		Issuer:         "test-issuer",
	}

	service := NewJWTService(config)

	claims := Claims{
		UserID:    "test-user-123",
		TokenType: string(AccessTokenType),
		Scopes:    []string{"read", "write"},
		Metadata:  map[string]interface{}{"role": "admin"},
	}

	token, err := service.GenerateAccessToken(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, token.Token)
	assert.True(t, token.IsAccessToken())
	assert.False(t, token.IsExpired())
	assert.Equal(t, "Bearer", token.TokenType)
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	config := Config{
		Secret:          "test-secret-key-123",
		RefreshTokenTTL: time.Hour * 24,
	}

	service := NewJWTService(config)

	claims := Claims{
		UserID:    "test-user-123",
		TokenType: string(RefreshTokenType),
	}

	token, err := service.GenerateRefreshToken(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, token.Token)
	assert.True(t, token.IsRefreshToken())
	assert.False(t, token.IsExpired())
}

func TestJWTService_GenerateTokenPair(t *testing.T) {
	config := Config{
		Secret:          "test-secret-key-123",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour * 24,
	}

	service := NewJWTService(config)

	claims := Claims{
		UserID:    "test-user-123",
		TokenType: string(AccessTokenType),
		Scopes:    []string{"read", "write"},
	}

	// Generate pair without refresh token
	pair, err := service.GenerateTokenPair(claims)
	require.NoError(t, err)
	assert.NotNil(t, pair.AccessToken)
	assert.Nil(t, pair.RefreshToken)
	assert.False(t, pair.HasRefreshToken())

	// Generate pair with refresh token
	pair, err = service.GenerateTokenPair(claims, func(cfg *TokenPairConfig) error {
		cfg.IncludeRefresh = true
		return nil
	})
	require.NoError(t, err)
	assert.NotNil(t, pair.AccessToken)
	assert.NotNil(t, pair.RefreshToken)
	assert.True(t, pair.HasRefreshToken())
}

func TestJWTService_ValidateToken(t *testing.T) {
	config := Config{
		Secret:         "test-secret-key-123",
		AccessTokenTTL: time.Hour,
	}

	service := NewJWTService(config)

	claims := Claims{
		UserID:    "test-user-123",
		TokenType: string(AccessTokenType),
		Scopes:    []string{"read", "write"},
	}

	// Generate a token
	token, err := service.GenerateAccessToken(claims)
	require.NoError(t, err)

	// Validate the token
	validated, err := service.ValidateToken(token.Token)
	require.NoError(t, err)
	assert.True(t, validated.IsValid)
	assert.Equal(t, claims.UserID, validated.Claims.UserID)
	assert.Equal(t, claims.TokenType, validated.Claims.TokenType)
	assert.Equal(t, claims.Scopes, validated.Claims.Scopes)
}

func TestJWTService_RefreshAccessToken(t *testing.T) {
	config := Config{
		Secret:          "test-secret-key-123",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour * 24,
	}

	service := NewJWTService(config)

	claims := Claims{
		UserID:    "test-user-123",
		TokenType: string(RefreshTokenType),
		Scopes:    []string{"read", "write"},
		Metadata:  map[string]interface{}{"role": "admin"},
	}

	// Generate refresh token
	refreshToken, err := service.GenerateRefreshToken(claims)
	require.NoError(t, err)

	// Refresh access token
	newAccessToken, err := service.RefreshAccessToken(refreshToken.Token)
	require.NoError(t, err)
	assert.NotNil(t, newAccessToken)
	assert.True(t, newAccessToken.IsAccessToken())
	assert.False(t, newAccessToken.IsExpired())
}

func TestJWTService_InvalidToken(t *testing.T) {
	config := Config{
		Secret: "test-secret-key-123",
	}

	service := NewJWTService(config)

	// Test with invalid token
	_, err := service.ValidateToken("invalid-token")
	assert.Error(t, err)
}

func TestJWTService_ExpiredToken(t *testing.T) {
	config := Config{
		Secret:         "test-secret-key-123",
		AccessTokenTTL: -time.Hour, // Already expired
	}

	service := NewJWTService(config)

	claims := Claims{
		UserID:    "test-user-123",
		TokenType: string(AccessTokenType),
	}

	token, err := service.GenerateAccessToken(claims)
	require.NoError(t, err)

	// Validate expired token
	_, err = service.ValidateToken(token.Token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestJWTService_Options(t *testing.T) {
	config := Config{
		Secret:         "test-secret-key-123",
		AccessTokenTTL: time.Hour,
	}

	service := NewJWTService(config)

	claims := Claims{
		UserID:    "test-user-123",
		TokenType: string(AccessTokenType),
	}

	// Test with custom options
	token, err := service.GenerateAccessToken(claims, func(cfg *TokenConfig) error {
		cfg.Subject = "custom-subject"
		cfg.Audience = []string{"custom-audience"}
		return nil
	})
	require.NoError(t, err)
	assert.NotEmpty(t, token.Token)

	// Validate and check claims
	validated, err := service.ValidateToken(token.Token)
	require.NoError(t, err)
	assert.Equal(t, "custom-subject", validated.Claims.Subject)
}

func TestClaimsParser_Component(t *testing.T) {
	config := Config{
		Secret: "test-secret-key-123",
	}

	parser := NewClaimsParser(config)

	// Create a test token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := jwt.MapClaims{
		"user_id":      "test-user-123",
		"token_type":   "access",
		"exp":          time.Now().Add(time.Hour).Unix(),
		"iat":          time.Now().Unix(),
		"nbf":          time.Now().Unix(),
		"scopes":       []interface{}{"read", "write"},
		"metadata":     map[string]interface{}{"role": "admin"},
		"custom_field": "custom_value",
	}
	token.Claims = claims

	// Test ParseClaims
	parsedClaims, err := parser.ParseClaims(token)
	require.NoError(t, err)
	assert.Equal(t, "test-user-123", parsedClaims.UserID)
	assert.Equal(t, "access", parsedClaims.TokenType)
	assert.Equal(t, []string{"read", "write"}, parsedClaims.Scopes)
	assert.Equal(t, "admin", parsedClaims.Metadata["role"])
	assert.Equal(t, "custom_value", parsedClaims.Custom["custom_field"])

	// Test ExtractClaims
	extractedClaims, err := parser.ExtractClaims(claims)
	require.NoError(t, err)
	assert.Equal(t, "test-user-123", extractedClaims.UserID)
	assert.Equal(t, "access", extractedClaims.TokenType)
}

func TestTokenGenerator_Component(t *testing.T) {
	config := Config{
		Secret:         "test-secret-key-123",
		AccessTokenTTL: time.Hour,
		Issuer:         "test-issuer",
	}

	parser := NewClaimsParser(config)
	generator := NewTokenGenerator(config, parser)

	claims := Claims{
		UserID:    "test-user-123",
		TokenType: string(AccessTokenType),
		Scopes:    []string{"read", "write"},
		Metadata:  map[string]interface{}{"role": "admin"},
		Custom:    map[string]interface{}{"custom_field": "custom_value"},
	}

	token, err := generator.GenerateAccessToken(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, token.Token)
	assert.True(t, token.IsAccessToken())
}

func TestTokenValidator_Component(t *testing.T) {
	config := Config{
		Secret:         "test-secret-key-123",
		AccessTokenTTL: time.Hour,
	}

	parser := NewClaimsParser(config)
	generator := NewTokenGenerator(config, parser)
	validator := NewTokenValidator(config, parser)

	claims := Claims{
		UserID:    "test-user-123",
		TokenType: string(AccessTokenType),
		Scopes:    []string{"read", "write"},
	}

	// Generate token
	token, err := generator.GenerateAccessToken(claims)
	require.NoError(t, err)

	// Validate token
	validated, err := validator.ValidateToken(token.Token)
	require.NoError(t, err)
	assert.True(t, validated.IsValid)
	assert.Equal(t, claims.UserID, validated.Claims.UserID)

	// Test claim validation
	validationConfig := &ValidationConfig{
		RequiredScopes:   []string{"read"},
		RequiredIssuer:   "test-issuer",
		RequiredAudience: "test-audience",
	}

	// This should fail because the token doesn't have the required audience
	err = validator.ValidateClaims(validated.Claims, validationConfig)
	assert.Error(t, err)
}

func TestTokenRefresher_Component(t *testing.T) {
	config := Config{
		Secret:          "test-secret-key-123",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour * 24,
	}

	parser := NewClaimsParser(config)
	generator := NewTokenGenerator(config, parser)
	validator := NewTokenValidator(config, parser)
	refresher := NewTokenRefresher(validator, generator)

	claims := Claims{
		UserID:    "test-user-123",
		TokenType: string(RefreshTokenType),
		Scopes:    []string{"read", "write"},
		Metadata:  map[string]interface{}{"role": "admin"},
	}

	// Generate refresh token
	refreshToken, err := generator.GenerateRefreshToken(claims)
	require.NoError(t, err)

	// Refresh access token
	newAccessToken, err := refresher.RefreshAccessToken(refreshToken.Token)
	require.NoError(t, err)
	assert.NotNil(t, newAccessToken)
	assert.True(t, newAccessToken.IsAccessToken())
}

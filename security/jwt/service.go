package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtService struct {
	config Config
}

// NewJWTService creates a new JWT service instance
func NewJWTService(config Config) Service {
	return &jwtService{
		config: config,
	}
}

func (s *jwtService) GenerateAccessToken(claims Claims, opts ...TokenOption) (*SignedToken, error) {
	config := &TokenConfig{
		TokenKind: AccessTokenType,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(s.config.AccessTokenTTL),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return s.generateSignedToken(claims, config)
}

func (s *jwtService) GenerateRefreshToken(claims Claims, opts ...RefreshOption) (*SignedToken, error) {
	config := &TokenConfig{
		TokenKind: RefreshTokenType,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(s.config.RefreshTokenTTL),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return s.generateSignedToken(claims, config)
}

func (s *jwtService) GenerateTokenPair(claims Claims, opts ...TokenPairOption) (*TokenPair, error) {
	pairConfig := &TokenPairConfig{
		IncludeRefresh: false,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(pairConfig); err != nil {
			return nil, err
		}
	}

	// Generate access token
	accessToken, err := s.GenerateAccessToken(claims, pairConfig.AccessTokenOptions...)
	if err != nil {
		return nil, err
	}

	var refreshToken *SignedToken
	if pairConfig.IncludeRefresh {
		refreshToken, err = s.GenerateRefreshToken(claims, pairConfig.RefreshTokenOptions...)
		if err != nil {
			return nil, err
		}
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *jwtService) ValidateToken(tokenString string, opts ...ValidationOption) (*ValidatedToken, error) {
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
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Parse claims
	parsedClaims, err := s.parseClaims(token)
	if err != nil {
		return nil, err
	}

	// Validate claims if not skipped
	if !config.SkipValidation {
		if err := s.validateClaims(parsedClaims, config); err != nil {
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

func (s *jwtService) RefreshAccessToken(refreshToken string, opts ...RefreshOption) (*SignedToken, error) {
	// Validate refresh token
	validated, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if !validated.IsValid {
		return nil, ErrInvalidToken
	}

	// Check if it's a refresh token
	if validated.Claims.TokenType != string(RefreshTokenType) {
		return nil, ErrInvalidToken
	}

	// Create new access token with same claims
	claims := Claims{
		UserID:    validated.Claims.UserID,
		TokenType: string(AccessTokenType),
		ClientID:  validated.Claims.ClientID,
		DeviceID:  validated.Claims.DeviceID,
		SessionID: validated.Claims.SessionID,
		Scopes:    validated.Claims.Scopes,
		Metadata:  validated.Claims.Metadata,
		Issuer:    validated.Claims.Issuer,
		Audience:  validated.Claims.Audience,
		Subject:   validated.Claims.Subject,
		Custom:    validated.Claims.Custom,
	}

	// Convert RefreshOption to TokenOption for access token generation
	tokenOpts := make([]TokenOption, len(opts))
	for i, opt := range opts {
		tokenOpts[i] = func(tc *TokenConfig) error {
			return opt(tc)
		}
	}

	return s.GenerateAccessToken(claims, tokenOpts...)
}

func (s *jwtService) generateSignedToken(claims Claims, config *TokenConfig) (*SignedToken, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	tokenClaims := token.Claims.(jwt.MapClaims)

	// Set standard claims
	tokenClaims["user_id"] = claims.UserID
	tokenClaims["token_type"] = string(config.TokenKind)
	tokenClaims["exp"] = config.ExpiresAt.Unix()
	tokenClaims["iat"] = config.IssuedAt.Unix()
	tokenClaims["nbf"] = config.IssuedAt.Unix()

	// Set optional claims
	if claims.ClientID != "" {
		tokenClaims["client_id"] = claims.ClientID
	}
	if claims.DeviceID != "" {
		tokenClaims["device_id"] = claims.DeviceID
	}
	if claims.SessionID != "" {
		tokenClaims["session_id"] = claims.SessionID
	}
	if len(claims.Scopes) > 0 {
		tokenClaims["scopes"] = claims.Scopes
	}
	if claims.Metadata != nil {
		tokenClaims["metadata"] = claims.Metadata
	}
	if claims.Issuer != "" {
		tokenClaims["iss"] = claims.Issuer
	} else if s.config.Issuer != "" {
		tokenClaims["iss"] = s.config.Issuer
	}
	if len(config.Audience) > 0 {
		tokenClaims["aud"] = config.Audience
	} else if claims.Audience != "" {
		tokenClaims["aud"] = claims.Audience
	} else if s.config.Audience != "" {
		tokenClaims["aud"] = s.config.Audience
	}
	if len(config.Scope) > 0 {
		tokenClaims["scope"] = config.Scope
	}
	if config.Subject != "" {
		tokenClaims["sub"] = config.Subject
	} else if claims.Subject != "" {
		tokenClaims["sub"] = claims.Subject
	} else {
		tokenClaims["sub"] = claims.UserID
	}

	// Add custom claims
	for key, value := range claims.Custom {
		tokenClaims[key] = value
	}

	// Sign token
	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return nil, err
	}

	return &SignedToken{
		Token:     tokenString,
		ExpiresAt: config.ExpiresAt,
		IssuedAt:  config.IssuedAt,
		TokenType: "Bearer",
		ExpiresIn: int64(config.ExpiresAt.Sub(config.IssuedAt).Seconds()),
		tokenKind: config.TokenKind,
		scope:     config.Scope,
	}, nil
}

func (s *jwtService) parseClaims(token *jwt.Token) (*Claims, error) {
	if token == nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidClaims
	}

	parsedClaims := &Claims{
		UserID:    userID,
		TokenType: s.getStringClaim(claims, "token_type"),
		ClientID:  s.getStringClaim(claims, "client_id"),
		DeviceID:  s.getStringClaim(claims, "device_id"),
		SessionID: s.getStringClaim(claims, "session_id"),
		Issuer:    s.getStringClaim(claims, "iss"),
		Audience:  s.getStringClaim(claims, "aud"),
		Subject:   s.getStringClaim(claims, "sub"),
		ExpiresAt: s.getInt64Claim(claims, "exp"),
		IssuedAt:  s.getInt64Claim(claims, "iat"),
		NotBefore: s.getInt64Claim(claims, "nbf"),
		Scopes:    s.getStringArrayClaim(claims, "scopes"),
		Metadata:  s.getMapClaim(claims, "metadata"),
		Custom:    s.getCustomClaims(claims),
	}

	return parsedClaims, nil
}

func (s *jwtService) validateClaims(claims *Claims, config *ValidationConfig) error {
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
		if !s.hasRequiredScopes(claims.Scopes, config.RequiredScopes) {
			return ErrInvalidToken
		}
	}

	return nil
}

func (s *jwtService) hasRequiredScopes(tokenScopes, requiredScopes []string) bool {
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

// Helper methods for claim parsing
func (s *jwtService) getStringClaim(claims jwt.MapClaims, key string) string {
	if value, ok := claims[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func (s *jwtService) getInt64Claim(claims jwt.MapClaims, key string) int64 {
	if value, ok := claims[key]; ok {
		switch v := value.(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case int:
			return int64(v)
		}
	}
	return 0
}

func (s *jwtService) getStringArrayClaim(claims jwt.MapClaims, key string) []string {
	if value, ok := claims[key]; ok {
		if arr, ok := value.([]interface{}); ok {
			var result []string
			for _, item := range arr {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return nil
}

func (s *jwtService) getMapClaim(claims jwt.MapClaims, key string) map[string]interface{} {
	if value, ok := claims[key]; ok {
		if m, ok := value.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

func (s *jwtService) getCustomClaims(claims jwt.MapClaims) map[string]interface{} {
	protectedKeys := map[string]bool{
		"user_id":    true,
		"token_type": true,
		"client_id":  true,
		"device_id":  true,
		"session_id": true,
		"scopes":     true,
		"metadata":   true,
		"iss":        true,
		"aud":        true,
		"sub":        true,
		"exp":        true,
		"iat":        true,
		"nbf":        true,
	}

	customClaims := make(map[string]interface{})
	for key, value := range claims {
		if !protectedKeys[key] {
			customClaims[key] = value
		}
	}
	return customClaims
}
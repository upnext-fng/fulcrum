package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type tokenGenerator struct {
	config       Config
	claimsParser ClaimsParser
}

func TokenGeneratorProvider(config Config, parser ClaimsParser) TokenGenerator {
	return NewTokenGenerator(config, parser)
}

func NewTokenGenerator(config Config, parser ClaimsParser) TokenGenerator {
	return &tokenGenerator{
		config:       config,
		claimsParser: parser,
	}
}

func (g *tokenGenerator) GenerateAccessToken(claims Claims, opts ...TokenOption) (*SignedToken, error) {
	config := &TokenConfig{
		TokenKind: AccessTokenType,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(g.config.AccessTokenTTL),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return g.generateSignedToken(claims, config)
}

func (g *tokenGenerator) GenerateRefreshToken(claims Claims, opts ...RefreshOption) (*SignedToken, error) {
	config := &TokenConfig{
		TokenKind: RefreshTokenType,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(g.config.RefreshTokenTTL),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return g.generateSignedToken(claims, config)
}

func (g *tokenGenerator) GenerateTokenPair(claims Claims, opts ...TokenPairOption) (*TokenPair, error) {
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
	accessToken, err := g.GenerateAccessToken(claims, pairConfig.AccessTokenOptions...)
	if err != nil {
		return nil, err
	}

	var refreshToken *SignedToken
	if pairConfig.IncludeRefresh {
		refreshToken, err = g.GenerateRefreshToken(claims, pairConfig.RefreshTokenOptions...)
		if err != nil {
			return nil, err
		}
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (g *tokenGenerator) generateSignedToken(claims Claims, config *TokenConfig) (*SignedToken, error) {
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
	} else if g.config.Issuer != "" {
		tokenClaims["iss"] = g.config.Issuer
	}
	if len(config.Audience) > 0 {
		tokenClaims["aud"] = config.Audience
	} else if claims.Audience != "" {
		tokenClaims["aud"] = claims.Audience
	} else if g.config.Audience != "" {
		tokenClaims["aud"] = g.config.Audience
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
	tokenString, err := token.SignedString([]byte(g.config.Secret))
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

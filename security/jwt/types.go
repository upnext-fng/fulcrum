package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Core token types
type TokenType string

const (
	AccessTokenType  TokenType = "access"
	RefreshTokenType TokenType = "refresh"
)

type SignedToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	TokenType string    `json:"token_type"` // "Bearer"
	ExpiresIn int64     `json:"expires_in"`

	// Internal metadata (không expose trong JSON)
	tokenKind TokenType `json:"-"`
	scope     []string  `json:"-"`
}

type TokenPair struct {
	AccessToken  *SignedToken `json:"access_token"`
	RefreshToken *SignedToken `json:"refresh_token,omitempty"`
}

// Configuration structures
type TokenConfig struct {
	TokenKind TokenType
	ExpiresAt time.Time
	IssuedAt  time.Time
	Audience  []string
	Scope     []string
	Subject   string
}

type TokenPairConfig struct {
	IncludeRefresh      bool
	AccessTokenOptions  []TokenOption
	RefreshTokenOptions []RefreshOption
}

// Option types
type TokenOption func(*TokenConfig) error
type RefreshOption func(*TokenConfig) error
type TokenPairOption func(*TokenPairConfig) error
type ValidationOption func(*ValidationConfig) error

type Claims struct {
	UserID    string                 `json:"user_id"`
	TokenType string                 `json:"token_type"`
	ClientID  string                 `json:"client_id,omitempty"`
	DeviceID  string                 `json:"device_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	Scopes    []string               `json:"scopes,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Issuer    string                 `json:"iss,omitempty"`
	Audience  string                 `json:"aud,omitempty"`
	Subject   string                 `json:"sub,omitempty"`
	ExpiresAt int64                  `json:"exp,omitempty"`
	IssuedAt  int64                  `json:"iat,omitempty"`
	NotBefore int64                  `json:"nbf,omitempty"`
	Custom    map[string]interface{} `json:"custom,omitempty"`
}

// Helper methods for Claims
func (c *Claims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

func (c *Claims) GetMetadataString(key string) string {
	if c.Metadata == nil {
		return ""
	}
	if value, exists := c.Metadata[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func (c *Claims) GetMetadataBool(key string) bool {
	if c.Metadata == nil {
		return false
	}
	if value, exists := c.Metadata[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

func (c *Claims) GetMetadataInt(key string) int {
	if c.Metadata == nil {
		return 0
	}
	if value, exists := c.Metadata[key]; exists {
		if i, ok := value.(float64); ok {
			return int(i)
		}
		if i, ok := value.(int); ok {
			return i
		}
	}
	return 0
}

func (c *Claims) GetCustomString(key string) string {
	if c.Custom == nil {
		return ""
	}
	if value, exists := c.Custom[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func (c *Claims) GetCustomBool(key string) bool {
	if c.Custom == nil {
		return false
	}
	if value, exists := c.Custom[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

func (c *Claims) SetMetadata(key string, value interface{}) {
	if c.Metadata == nil {
		c.Metadata = make(map[string]interface{})
	}
	c.Metadata[key] = value
}

func (c *Claims) SetCustom(key string, value interface{}) {
	if c.Custom == nil {
		c.Custom = make(map[string]interface{})
	}
	c.Custom[key] = value
}

func (c *Claims) IsExpired() bool {
	if c.ExpiresAt == 0 {
		return false
	}
	return time.Now().Unix() > c.ExpiresAt
}

func (c *Claims) IsValidAt(t time.Time) bool {
	timestamp := t.Unix()

	// Check not before
	if c.NotBefore > 0 && timestamp < c.NotBefore {
		return false
	}

	// Check expiration
	if c.ExpiresAt > 0 && timestamp > c.ExpiresAt {
		return false
	}

	return true
}

type ValidatedToken struct {
	Token     *jwt.Token `json:"-"`
	Claims    *Claims    `json:"claims"`
	IsValid   bool       `json:"is_valid"`
	ExpiresAt time.Time  `json:"expires_at"`
}

type ValidationConfig struct {
	SkipExpiration   bool
	SkipValidation   bool
	RequiredScopes   []string
	RequiredIssuer   string
	RequiredAudience string
}

// Methods cho SignedToken
func (st *SignedToken) IsExpired() bool {
	return time.Now().After(st.ExpiresAt)
}

func (st *SignedToken) AuthHeader() string {
	return fmt.Sprintf("%s %s", st.TokenType, st.Token)
}

func (st *SignedToken) ShouldRefresh(threshold time.Duration) bool {
	return time.Until(st.ExpiresAt) <= threshold
}

func (st *SignedToken) IsAccessToken() bool {
	return st.tokenKind == AccessTokenType
}

func (st *SignedToken) IsRefreshToken() bool {
	return st.tokenKind == RefreshTokenType
}

// Methods cho TokenPair
func (tp *TokenPair) HasRefreshToken() bool {
	return tp.RefreshToken != nil
}

func (tp *TokenPair) BothValid() bool {
	if tp.AccessToken == nil || tp.AccessToken.IsExpired() {
		return false
	}

	if tp.RefreshToken != nil && tp.RefreshToken.IsExpired() {
		return false
	}

	return true
}

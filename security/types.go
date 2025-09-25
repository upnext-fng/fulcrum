package security

import (
	"time"
)

// UserClaims represents strongly-typed user authentication data
type UserClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
	
	// Optional fields
	FirstName   string    `json:"first_name,omitempty"`
	LastName    string    `json:"last_name,omitempty"`
	DisplayName string    `json:"display_name,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Timezone    string    `json:"timezone,omitempty"`
	Locale      string    `json:"locale,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	
	// Permission and access control
	Permissions []string `json:"permissions,omitempty"`
	Groups      []string `json:"groups,omitempty"`
	
	// Organization/tenant context
	OrganizationID string `json:"organization_id,omitempty"`
	TenantID       string `json:"tenant_id,omitempty"`
	
	// Session context
	SessionID string `json:"session_id,omitempty"`
	DeviceID  string `json:"device_id,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// TokenMetadata represents strongly-typed token metadata
type TokenMetadata struct {
	// Token context
	Purpose     string    `json:"purpose,omitempty"`      // "login", "api", "refresh", etc.
	Source      string    `json:"source,omitempty"`       // "web", "mobile", "api", etc.
	Environment string    `json:"environment,omitempty"`  // "production", "staging", "development"
	
	// Security context
	RequiresMFA bool     `json:"requires_mfa,omitempty"`
	MFAVerified bool     `json:"mfa_verified,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	
	// Audit and tracking
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by,omitempty"`
	RequestID   string    `json:"request_id,omitempty"`
	TraceID     string    `json:"trace_id,omitempty"`
	
	// Rate limiting and throttling
	RateLimit     int `json:"rate_limit,omitempty"`
	RequestsUsed  int `json:"requests_used,omitempty"`
	
	// Feature flags and experiments
	FeatureFlags map[string]bool `json:"feature_flags,omitempty"`
	Experiments  map[string]string `json:"experiments,omitempty"`
}

// CustomClaims represents application-specific custom claims
type CustomClaims struct {
	// Application-specific data
	AppVersion    string            `json:"app_version,omitempty"`
	APIVersion    string            `json:"api_version,omitempty"`
	
	// Business logic data
	SubscriptionTier   string    `json:"subscription_tier,omitempty"`
	SubscriptionExpiry *time.Time `json:"subscription_expiry,omitempty"`
	
	// Preferences
	Theme         string            `json:"theme,omitempty"`
	Language      string            `json:"language,omitempty"`
	Preferences   map[string]string `json:"preferences,omitempty"`
	
	// Integration data
	ExternalIDs   map[string]string `json:"external_ids,omitempty"`
	
	// Extensible custom data for specific use cases
	Data map[string]interface{} `json:"data,omitempty"`
}

// TokenRequest represents a strongly-typed token generation request
type TokenRequest struct {
	UserClaims    UserClaims    `json:"user_claims"`
	Metadata      TokenMetadata `json:"metadata"`
	CustomClaims  CustomClaims  `json:"custom_claims"`
	
	// Token configuration
	ExpiresIn     *time.Duration `json:"expires_in,omitempty"`
	RefreshToken  bool          `json:"refresh_token,omitempty"`
	Audience      []string      `json:"audience,omitempty"`
	Issuer        string        `json:"issuer,omitempty"`
}

// TokenResponse represents a strongly-typed token generation response
type TokenResponse struct {
	AccessToken  string     `json:"access_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int64      `json:"expires_in"`
	ExpiresAt    time.Time  `json:"expires_at"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	Scope        []string   `json:"scope,omitempty"`
}

// ValidationRequest represents token validation parameters
type ValidationRequest struct {
	Token            string   `json:"token"`
	RequiredScopes   []string `json:"required_scopes,omitempty"`
	RequiredAudience string   `json:"required_audience,omitempty"`
	RequiredIssuer   string   `json:"required_issuer,omitempty"`
	SkipExpiration   bool     `json:"skip_expiration,omitempty"`
}

// ValidationResponse represents token validation results
type ValidationResponse struct {
	Valid        bool          `json:"valid"`
	UserClaims   UserClaims    `json:"user_claims"`
	Metadata     TokenMetadata `json:"metadata"`
	CustomClaims CustomClaims  `json:"custom_claims"`
	ExpiresAt    time.Time     `json:"expires_at"`
	IssuedAt     time.Time     `json:"issued_at"`
	Scopes       []string      `json:"scopes,omitempty"`
}

// Helper methods for UserClaims
func (uc *UserClaims) HasPermission(permission string) bool {
	for _, p := range uc.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func (uc *UserClaims) HasRole(role string) bool {
	return uc.Role == role
}

func (uc *UserClaims) InGroup(group string) bool {
	for _, g := range uc.Groups {
		if g == group {
			return true
		}
	}
	return false
}

func (uc *UserClaims) GetDisplayName() string {
	if uc.DisplayName != "" {
		return uc.DisplayName
	}
	if uc.FirstName != "" && uc.LastName != "" {
		return uc.FirstName + " " + uc.LastName
	}
	if uc.Username != "" {
		return uc.Username
	}
	return uc.UserID
}

// Helper methods for TokenMetadata
func (tm *TokenMetadata) HasScope(scope string) bool {
	for _, s := range tm.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

func (tm *TokenMetadata) IsFeatureEnabled(feature string) bool {
	if tm.FeatureFlags == nil {
		return false
	}
	return tm.FeatureFlags[feature]
}

// Helper methods for CustomClaims
func (cc *CustomClaims) GetPreference(key string) string {
	if cc.Preferences == nil {
		return ""
	}
	return cc.Preferences[key]
}

func (cc *CustomClaims) GetExternalID(provider string) string {
	if cc.ExternalIDs == nil {
		return ""
	}
	return cc.ExternalIDs[provider]
}

// Validation methods
func (tr *TokenRequest) Validate() error {
	if tr.UserClaims.UserID == "" {
		return ErrInvalidUserID
	}
	return nil
}

func (vr *ValidationRequest) Validate() error {
	if vr.Token == "" {
		return ErrInvalidToken
	}
	return nil
}

// Error definitions
var (
	ErrInvalidUserID    = fmt.Errorf("invalid user ID")
	ErrInvalidToken     = fmt.Errorf("invalid token")
	ErrInvalidRequest   = fmt.Errorf("invalid request")
	ErrUnauthorized     = fmt.Errorf("unauthorized")
	ErrTokenExpired     = fmt.Errorf("token expired")
	ErrInvalidScope     = fmt.Errorf("invalid scope")
	ErrInvalidAudience  = fmt.Errorf("invalid audience")
	ErrInvalidIssuer    = fmt.Errorf("invalid issuer")
)

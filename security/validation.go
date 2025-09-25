package security

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

// Validation constants
const (
	MinUserIDLength = 1
	MaxUserIDLength = 255
	MaxUsernameLength = 100
	MaxEmailLength = 320
	MaxRoleLength = 100
	MaxNameLength = 255
	MaxURLLength = 2048
	MaxTimezoneLength = 50
	MaxLocaleLength = 10
	MaxPermissionLength = 100
	MaxGroupLength = 100
	MaxOrganizationIDLength = 255
	MaxTenantIDLength = 255
	MaxSessionIDLength = 255
	MaxDeviceIDLength = 255
	MaxClientIDLength = 255
	MaxIPAddressLength = 45 // IPv6 max length
	MaxUserAgentLength = 1000
	MaxPurposeLength = 50
	MaxSourceLength = 50
	MaxEnvironmentLength = 50
	MaxRequestIDLength = 255
	MaxTraceIDLength = 255
	MaxAppVersionLength = 50
	MaxAPIVersionLength = 50
	MaxSubscriptionTierLength = 50
	MaxThemeLength = 50
	MaxLanguageLength = 10
	MaxPreferenceKeyLength = 100
	MaxPreferenceValueLength = 1000
	MaxExternalIDKeyLength = 100
	MaxExternalIDValueLength = 255
	MaxAudienceLength = 255
	MaxIssuerLength = 255
	MaxScopeLength = 100
	MaxFeatureFlagLength = 100
	MaxExperimentKeyLength = 100
	MaxExperimentValueLength = 255
)

// Regular expressions for validation
var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	roleRegex     = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	scopeRegex    = regexp.MustCompile(`^[a-zA-Z0-9_:.-]+$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", ve.Field, ve.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}
	if len(ve) == 1 {
		return ve[0].Error()
	}
	
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("multiple validation errors: %s", strings.Join(messages, "; "))
}

func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Validator provides validation functionality
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string, value interface{}) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// ValidateString validates a string field
func (v *Validator) ValidateString(field, value string, required bool, minLen, maxLen int) {
	if required && value == "" {
		v.AddError(field, "is required", value)
		return
	}
	
	if value != "" {
		if len(value) < minLen {
			v.AddError(field, fmt.Sprintf("must be at least %d characters", minLen), value)
		}
		if len(value) > maxLen {
			v.AddError(field, fmt.Sprintf("must not exceed %d characters", maxLen), value)
		}
	}
}

// ValidateEmail validates an email address
func (v *Validator) ValidateEmail(field, email string, required bool) {
	if required && email == "" {
		v.AddError(field, "is required", email)
		return
	}
	
	if email != "" {
		if len(email) > MaxEmailLength {
			v.AddError(field, fmt.Sprintf("must not exceed %d characters", MaxEmailLength), email)
			return
		}
		
		if _, err := mail.ParseAddress(email); err != nil {
			v.AddError(field, "must be a valid email address", email)
		}
	}
}

// ValidateUsername validates a username
func (v *Validator) ValidateUsername(field, username string, required bool) {
	v.ValidateString(field, username, required, 1, MaxUsernameLength)
	
	if username != "" && !usernameRegex.MatchString(username) {
		v.AddError(field, "must contain only letters, numbers, dots, hyphens, and underscores", username)
	}
}

// ValidateRole validates a role
func (v *Validator) ValidateRole(field, role string, required bool) {
	v.ValidateString(field, role, required, 1, MaxRoleLength)
	
	if role != "" && !roleRegex.MatchString(role) {
		v.AddError(field, "must contain only letters, numbers, dots, hyphens, and underscores", role)
	}
}

// ValidateURL validates a URL
func (v *Validator) ValidateURL(field, url string, required bool) {
	v.ValidateString(field, url, required, 1, MaxURLLength)
	
	if url != "" {
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			v.AddError(field, "must be a valid HTTP or HTTPS URL", url)
		}
	}
}

// ValidateStringSlice validates a slice of strings
func (v *Validator) ValidateStringSlice(field string, slice []string, required bool, maxItems int, itemMaxLen int) {
	if required && len(slice) == 0 {
		v.AddError(field, "is required", slice)
		return
	}
	
	if len(slice) > maxItems {
		v.AddError(field, fmt.Sprintf("must not exceed %d items", maxItems), slice)
	}
	
	for i, item := range slice {
		if len(item) > itemMaxLen {
			v.AddError(fmt.Sprintf("%s[%d]", field, i), fmt.Sprintf("must not exceed %d characters", itemMaxLen), item)
		}
	}
}

// ValidateScopes validates scopes
func (v *Validator) ValidateScopes(field string, scopes []string, required bool) {
	v.ValidateStringSlice(field, scopes, required, 50, MaxScopeLength)
	
	for i, scope := range scopes {
		if !scopeRegex.MatchString(scope) {
			v.AddError(fmt.Sprintf("%s[%d]", field, i), "must contain only letters, numbers, colons, dots, hyphens, and underscores", scope)
		}
	}
}

// ValidateUUID validates a UUID
func (v *Validator) ValidateUUID(field, uuid string, required bool) {
	if required && uuid == "" {
		v.AddError(field, "is required", uuid)
		return
	}
	
	if uuid != "" && !uuidRegex.MatchString(strings.ToLower(uuid)) {
		v.AddError(field, "must be a valid UUID", uuid)
	}
}

// ValidateTime validates a time pointer
func (v *Validator) ValidateTime(field string, t *time.Time, required bool) {
	if required && (t == nil || t.IsZero()) {
		v.AddError(field, "is required", t)
	}
}

// ValidateDuration validates a duration pointer
func (v *Validator) ValidateDuration(field string, d *time.Duration, required bool, min, max time.Duration) {
	if required && d == nil {
		v.AddError(field, "is required", d)
		return
	}
	
	if d != nil {
		if *d < min {
			v.AddError(field, fmt.Sprintf("must be at least %v", min), *d)
		}
		if *d > max {
			v.AddError(field, fmt.Sprintf("must not exceed %v", max), *d)
		}
	}
}

// ValidateUserClaims validates UserClaims
func ValidateUserClaims(claims UserClaims) error {
	v := NewValidator()
	
	v.ValidateString("user_id", claims.UserID, true, MinUserIDLength, MaxUserIDLength)
	v.ValidateUsername("username", claims.Username, false)
	v.ValidateEmail("email", claims.Email, false)
	v.ValidateRole("role", claims.Role, false)
	v.ValidateString("first_name", claims.FirstName, false, 1, MaxNameLength)
	v.ValidateString("last_name", claims.LastName, false, 1, MaxNameLength)
	v.ValidateString("display_name", claims.DisplayName, false, 1, MaxNameLength)
	v.ValidateURL("avatar_url", claims.AvatarURL, false)
	v.ValidateString("timezone", claims.Timezone, false, 1, MaxTimezoneLength)
	v.ValidateString("locale", claims.Locale, false, 1, MaxLocaleLength)
	v.ValidateStringSlice("permissions", claims.Permissions, false, 100, MaxPermissionLength)
	v.ValidateStringSlice("groups", claims.Groups, false, 50, MaxGroupLength)
	v.ValidateString("organization_id", claims.OrganizationID, false, 1, MaxOrganizationIDLength)
	v.ValidateString("tenant_id", claims.TenantID, false, 1, MaxTenantIDLength)
	v.ValidateString("session_id", claims.SessionID, false, 1, MaxSessionIDLength)
	v.ValidateString("device_id", claims.DeviceID, false, 1, MaxDeviceIDLength)
	v.ValidateString("client_id", claims.ClientID, false, 1, MaxClientIDLength)
	v.ValidateString("ip_address", claims.IPAddress, false, 1, MaxIPAddressLength)
	v.ValidateString("user_agent", claims.UserAgent, false, 1, MaxUserAgentLength)
	
	if v.HasErrors() {
		return v.Errors()
	}
	return nil
}

// ValidateTokenMetadata validates TokenMetadata
func ValidateTokenMetadata(metadata TokenMetadata) error {
	v := NewValidator()
	
	v.ValidateString("purpose", metadata.Purpose, false, 1, MaxPurposeLength)
	v.ValidateString("source", metadata.Source, false, 1, MaxSourceLength)
	v.ValidateString("environment", metadata.Environment, false, 1, MaxEnvironmentLength)
	v.ValidateString("created_by", metadata.CreatedBy, false, 1, MaxUserIDLength)
	v.ValidateString("request_id", metadata.RequestID, false, 1, MaxRequestIDLength)
	v.ValidateString("trace_id", metadata.TraceID, false, 1, MaxTraceIDLength)
	v.ValidateScopes("scopes", metadata.Scopes, false)
	
	if metadata.RateLimit < 0 {
		v.AddError("rate_limit", "must be non-negative", metadata.RateLimit)
	}
	if metadata.RequestsUsed < 0 {
		v.AddError("requests_used", "must be non-negative", metadata.RequestsUsed)
	}
	
	if v.HasErrors() {
		return v.Errors()
	}
	return nil
}

// ValidateCustomClaims validates CustomClaims
func ValidateCustomClaims(claims CustomClaims) error {
	v := NewValidator()
	
	v.ValidateString("app_version", claims.AppVersion, false, 1, MaxAppVersionLength)
	v.ValidateString("api_version", claims.APIVersion, false, 1, MaxAPIVersionLength)
	v.ValidateString("subscription_tier", claims.SubscriptionTier, false, 1, MaxSubscriptionTierLength)
	v.ValidateString("theme", claims.Theme, false, 1, MaxThemeLength)
	v.ValidateString("language", claims.Language, false, 1, MaxLanguageLength)
	
	// Validate preferences map
	for key, value := range claims.Preferences {
		if len(key) > MaxPreferenceKeyLength {
			v.AddError(fmt.Sprintf("preferences[%s]", key), fmt.Sprintf("key must not exceed %d characters", MaxPreferenceKeyLength), key)
		}
		if len(value) > MaxPreferenceValueLength {
			v.AddError(fmt.Sprintf("preferences[%s]", key), fmt.Sprintf("value must not exceed %d characters", MaxPreferenceValueLength), value)
		}
	}
	
	// Validate external IDs map
	for key, value := range claims.ExternalIDs {
		if len(key) > MaxExternalIDKeyLength {
			v.AddError(fmt.Sprintf("external_ids[%s]", key), fmt.Sprintf("key must not exceed %d characters", MaxExternalIDKeyLength), key)
		}
		if len(value) > MaxExternalIDValueLength {
			v.AddError(fmt.Sprintf("external_ids[%s]", key), fmt.Sprintf("value must not exceed %d characters", MaxExternalIDValueLength), value)
		}
	}
	
	if v.HasErrors() {
		return v.Errors()
	}
	return nil
}

// ValidateTokenRequest validates TokenRequest
func ValidateTokenRequest(request TokenRequest) error {
	v := NewValidator()
	
	// Validate user claims
	if err := ValidateUserClaims(request.UserClaims); err != nil {
		if validationErrors, ok := err.(ValidationErrors); ok {
			v.errors = append(v.errors, validationErrors...)
		} else {
			v.AddError("user_claims", err.Error(), request.UserClaims)
		}
	}
	
	// Validate metadata
	if err := ValidateTokenMetadata(request.Metadata); err != nil {
		if validationErrors, ok := err.(ValidationErrors); ok {
			v.errors = append(v.errors, validationErrors...)
		} else {
			v.AddError("metadata", err.Error(), request.Metadata)
		}
	}
	
	// Validate custom claims
	if err := ValidateCustomClaims(request.CustomClaims); err != nil {
		if validationErrors, ok := err.(ValidationErrors); ok {
			v.errors = append(v.errors, validationErrors...)
		} else {
			v.AddError("custom_claims", err.Error(), request.CustomClaims)
		}
	}
	
	// Validate token configuration
	v.ValidateDuration("expires_in", request.ExpiresIn, false, time.Minute, time.Hour*24*365)
	v.ValidateStringSlice("audience", request.Audience, false, 10, MaxAudienceLength)
	v.ValidateString("issuer", request.Issuer, false, 1, MaxIssuerLength)
	
	if v.HasErrors() {
		return v.Errors()
	}
	return nil
}

// ValidateValidationRequest validates ValidationRequest
func ValidateValidationRequest(request ValidationRequest) error {
	v := NewValidator()
	
	v.ValidateString("token", request.Token, true, 1, 10000) // JWT tokens can be quite long
	v.ValidateScopes("required_scopes", request.RequiredScopes, false)
	v.ValidateString("required_audience", request.RequiredAudience, false, 1, MaxAudienceLength)
	v.ValidateString("required_issuer", request.RequiredIssuer, false, 1, MaxIssuerLength)
	
	if v.HasErrors() {
		return v.Errors()
	}
	return nil
}

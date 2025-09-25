package security

import (
	"time"

	"github.com/labstack/echo/v4"
	jwtmod "github.com/upnext-fng/fulcrum/security/jwt"
	"github.com/upnext-fng/fulcrum/security/middleware"
	"github.com/upnext-fng/fulcrum/security/password"
)

type manager struct {
	config            Config
	jwtService        jwtmod.Service
	claimsParser      jwtmod.ClaimsParser
	passwordService   password.Service
	middlewareService middleware.Service
}

func NewManager(config Config) SecurityService {
	jwtService := jwtmod.NewJWTService(config.JWT)

	// Extract claims parser from JWT service
	claimsParser := jwtmod.NewClaimsParser(config.JWT)

	// Update middleware config to include JWT config
	middlewareConfig := config.Middleware
	middlewareConfig.JWTConfig = config.JWT

	return &manager{
		config:            config,
		jwtService:        jwtService,
		claimsParser:      claimsParser,
		passwordService:   password.NewService(config.Password),
		middlewareService: middleware.NewService(middlewareConfig, jwtService),
	}
}

// GenerateToken generates a token using strongly-typed request
func (m *manager) GenerateToken(request TokenRequest) (TokenResponse, error) {
	// Validate request
	if err := request.Validate(); err != nil {
		return TokenResponse{}, err
	}

	// Convert to JWT claims
	jwtClaims, err := convertTokenRequestToJWTClaims(request)
	if err != nil {
		return TokenResponse{}, err
	}

	// Generate access token
	signedToken, err := m.jwtService.GenerateAccessToken(jwtClaims)
	if err != nil {
		return TokenResponse{}, err
	}

	// Build response
	response := TokenResponse{
		AccessToken: signedToken.Token,
		TokenType:   signedToken.TokenType,
		ExpiresIn:   signedToken.ExpiresIn,
		ExpiresAt:   signedToken.ExpiresAt,
	}

	// Generate refresh token if requested
	if request.RefreshToken {
		refreshToken, err := m.jwtService.GenerateRefreshToken(jwtClaims)
		if err != nil {
			return TokenResponse{}, err
		}
		response.RefreshToken = refreshToken.Token
	}

	// Set scopes
	if len(request.Metadata.Scopes) > 0 {
		response.Scope = request.Metadata.Scopes
	}

	return response, nil
}

// ValidateToken validates a token using strongly-typed request
func (m *manager) ValidateToken(request ValidationRequest) (ValidationResponse, error) {
	// Validate request
	if err := request.Validate(); err != nil {
		return ValidationResponse{}, err
	}

	// Validate token
	validatedToken, err := m.jwtService.ValidateToken(request.Token)
	if err != nil {
		return ValidationResponse{Valid: false}, err
	}

	// Parse claims
	claims, err := m.claimsParser.ParseClaims(validatedToken.Token)
	if err != nil {
		return ValidationResponse{Valid: false}, err
	}

	// Convert to response
	response, err := convertJWTClaimsToValidationResponse(claims)
	if err != nil {
		return ValidationResponse{Valid: false}, err
	}

	// Validate required scopes
	if len(request.RequiredScopes) > 0 {
		for _, requiredScope := range request.RequiredScopes {
			if !response.Metadata.HasScope(requiredScope) {
				return ValidationResponse{Valid: false}, ErrInvalidScope
			}
		}
	}

	// Validate required audience
	if request.RequiredAudience != "" {
		// Check if token has required audience
		// This would need to be implemented based on your JWT claims structure
	}

	// Validate required issuer
	if request.RequiredIssuer != "" {
		// Check if token has required issuer
		// This would need to be implemented based on your JWT claims structure
	}

	return response, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func (m *manager) RefreshAccessToken(refreshToken string) (TokenResponse, error) {
	// Use JWT service to refresh the access token
	signedToken, err := m.jwtService.RefreshAccessToken(refreshToken)
	if err != nil {
		return TokenResponse{}, err
	}

	// Build response
	response := TokenResponse{
		AccessToken: signedToken.Token,
		TokenType:   signedToken.TokenType,
		ExpiresIn:   signedToken.ExpiresIn,
		ExpiresAt:   signedToken.ExpiresAt,
	}

	return response, nil
}

func (m *manager) HashPassword(password string) (string, error) {
	return m.passwordService.HashPassword(password)
}

func (m *manager) VerifyPassword(hashedPassword, password string) error {
	return m.passwordService.VerifyPassword(hashedPassword, password)
}

func (m *manager) ValidatePassword(password string) error {
	return m.passwordService.ValidatePassword(password)
}

func (m *manager) JWTMiddleware() echo.MiddlewareFunc {
	return m.middlewareService.JWTMiddleware()
}

func (m *manager) AuthMiddleware() echo.MiddlewareFunc {
	return m.middlewareService.AuthMiddleware()
}

func (m *manager) CORSMiddleware() echo.MiddlewareFunc {
	return m.middlewareService.CORSMiddleware()
}

func (m *manager) RateLimitMiddleware() echo.MiddlewareFunc {
	return m.middlewareService.RateLimitMiddleware()
}

// Helper functions for conversion

// convertTokenRequestToJWTClaims converts TokenRequest to JWT Claims
func convertTokenRequestToJWTClaims(request TokenRequest) (jwtmod.Claims, error) {
	claims := jwtmod.Claims{
		UserID: request.UserClaims.UserID,
	}

	// Set metadata
	if claims.Metadata == nil {
		claims.Metadata = make(map[string]interface{})
	}

	// Add user claims to metadata
	claims.Metadata["username"] = request.UserClaims.Username
	claims.Metadata["email"] = request.UserClaims.Email
	claims.Metadata["role"] = request.UserClaims.Role
	claims.Metadata["first_name"] = request.UserClaims.FirstName
	claims.Metadata["last_name"] = request.UserClaims.LastName
	claims.Metadata["display_name"] = request.UserClaims.DisplayName
	claims.Metadata["avatar_url"] = request.UserClaims.AvatarURL
	claims.Metadata["timezone"] = request.UserClaims.Timezone
	claims.Metadata["locale"] = request.UserClaims.Locale
	claims.Metadata["permissions"] = request.UserClaims.Permissions
	claims.Metadata["groups"] = request.UserClaims.Groups
	claims.Metadata["organization_id"] = request.UserClaims.OrganizationID
	claims.Metadata["tenant_id"] = request.UserClaims.TenantID
	claims.Metadata["ip_address"] = request.UserClaims.IPAddress
	claims.Metadata["user_agent"] = request.UserClaims.UserAgent

	// Add token metadata
	claims.Metadata["purpose"] = request.Metadata.Purpose
	claims.Metadata["source"] = request.Metadata.Source
	claims.Metadata["environment"] = request.Metadata.Environment
	claims.Metadata["created_at"] = request.Metadata.CreatedAt
	claims.Metadata["created_by"] = request.Metadata.CreatedBy
	claims.Metadata["request_id"] = request.Metadata.RequestID
	claims.Metadata["trace_id"] = request.Metadata.TraceID
	claims.Metadata["requires_mfa"] = request.Metadata.RequiresMFA
	claims.Metadata["mfa_verified"] = request.Metadata.MFAVerified
	claims.Metadata["rate_limit"] = request.Metadata.RateLimit
	claims.Metadata["requests_used"] = request.Metadata.RequestsUsed
	claims.Metadata["feature_flags"] = request.Metadata.FeatureFlags
	claims.Metadata["experiments"] = request.Metadata.Experiments

	// Set custom claims
	if claims.Custom == nil {
		claims.Custom = make(map[string]interface{})
	}
	claims.Custom["app_version"] = request.CustomClaims.AppVersion
	claims.Custom["api_version"] = request.CustomClaims.APIVersion
	claims.Custom["subscription_tier"] = request.CustomClaims.SubscriptionTier
	claims.Custom["subscription_expiry"] = request.CustomClaims.SubscriptionExpiry
	claims.Custom["theme"] = request.CustomClaims.Theme
	claims.Custom["language"] = request.CustomClaims.Language
	claims.Custom["preferences"] = request.CustomClaims.Preferences
	claims.Custom["external_ids"] = request.CustomClaims.ExternalIDs
	claims.Custom["data"] = request.CustomClaims.Data

	// Set other fields
	if len(request.Metadata.Scopes) > 0 {
		claims.Scopes = request.Metadata.Scopes
	}
	if request.UserClaims.SessionID != "" {
		claims.SessionID = request.UserClaims.SessionID
	}
	if request.UserClaims.DeviceID != "" {
		claims.DeviceID = request.UserClaims.DeviceID
	}
	if request.UserClaims.ClientID != "" {
		claims.ClientID = request.UserClaims.ClientID
	}

	return claims, nil
}

// convertJWTClaimsToValidationResponse converts JWT Claims to ValidationResponse
func convertJWTClaimsToValidationResponse(claims *jwtmod.Claims) (ValidationResponse, error) {
	var response ValidationResponse

	response.Valid = true
	response.UserClaims.UserID = claims.UserID

	// Extract user claims from metadata
	if claims.Metadata != nil {
		if username, ok := claims.Metadata["username"].(string); ok {
			response.UserClaims.Username = username
		}
		if email, ok := claims.Metadata["email"].(string); ok {
			response.UserClaims.Email = email
		}
		if role, ok := claims.Metadata["role"].(string); ok {
			response.UserClaims.Role = role
		}
		if firstName, ok := claims.Metadata["first_name"].(string); ok {
			response.UserClaims.FirstName = firstName
		}
		if lastName, ok := claims.Metadata["last_name"].(string); ok {
			response.UserClaims.LastName = lastName
		}
		if displayName, ok := claims.Metadata["display_name"].(string); ok {
			response.UserClaims.DisplayName = displayName
		}
		if avatarURL, ok := claims.Metadata["avatar_url"].(string); ok {
			response.UserClaims.AvatarURL = avatarURL
		}
		if timezone, ok := claims.Metadata["timezone"].(string); ok {
			response.UserClaims.Timezone = timezone
		}
		if locale, ok := claims.Metadata["locale"].(string); ok {
			response.UserClaims.Locale = locale
		}
		if permissions, ok := claims.Metadata["permissions"].([]interface{}); ok {
			for _, p := range permissions {
				if perm, ok := p.(string); ok {
					response.UserClaims.Permissions = append(response.UserClaims.Permissions, perm)
				}
			}
		}
		if groups, ok := claims.Metadata["groups"].([]interface{}); ok {
			for _, g := range groups {
				if group, ok := g.(string); ok {
					response.UserClaims.Groups = append(response.UserClaims.Groups, group)
				}
			}
		}
		if orgID, ok := claims.Metadata["organization_id"].(string); ok {
			response.UserClaims.OrganizationID = orgID
		}
		if tenantID, ok := claims.Metadata["tenant_id"].(string); ok {
			response.UserClaims.TenantID = tenantID
		}
		if ipAddress, ok := claims.Metadata["ip_address"].(string); ok {
			response.UserClaims.IPAddress = ipAddress
		}
		if userAgent, ok := claims.Metadata["user_agent"].(string); ok {
			response.UserClaims.UserAgent = userAgent
		}

		// Extract metadata
		if purpose, ok := claims.Metadata["purpose"].(string); ok {
			response.Metadata.Purpose = purpose
		}
		if source, ok := claims.Metadata["source"].(string); ok {
			response.Metadata.Source = source
		}
		if environment, ok := claims.Metadata["environment"].(string); ok {
			response.Metadata.Environment = environment
		}
		if requiresMFA, ok := claims.Metadata["requires_mfa"].(bool); ok {
			response.Metadata.RequiresMFA = requiresMFA
		}
		if mfaVerified, ok := claims.Metadata["mfa_verified"].(bool); ok {
			response.Metadata.MFAVerified = mfaVerified
		}
	}

	// Set session info
	response.UserClaims.SessionID = claims.SessionID
	response.UserClaims.DeviceID = claims.DeviceID
	response.UserClaims.ClientID = claims.ClientID

	// Set scopes
	response.Scopes = claims.Scopes
	response.Metadata.Scopes = claims.Scopes

	// Set timestamps
	if claims.ExpiresAt > 0 {
		response.ExpiresAt = time.Unix(claims.ExpiresAt, 0)
	}
	if claims.IssuedAt > 0 {
		response.IssuedAt = time.Unix(claims.IssuedAt, 0)
	}

	return response, nil
}

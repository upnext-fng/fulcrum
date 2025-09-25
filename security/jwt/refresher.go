package jwt

type tokenRefresher struct {
	validator TokenValidator
	generator TokenGenerator
}

func TokenRefresherProvider(validator TokenValidator, generator TokenGenerator) TokenRefresher {
	return NewTokenRefresher(validator, generator)
}

func NewTokenRefresher(validator TokenValidator, generator TokenGenerator) TokenRefresher {
	return &tokenRefresher{
		validator: validator,
		generator: generator,
	}
}

func (r *tokenRefresher) RefreshAccessToken(refreshToken string, opts ...RefreshOption) (*SignedToken, error) {
	// Validate refresh token
	validated, err := r.validator.ValidateToken(refreshToken)
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

	return r.generator.GenerateAccessToken(claims, tokenOpts...)
}

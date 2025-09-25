package jwt

import "time"

// Token options
func WithAudience(audience ...string) TokenOption {
	return func(tc *TokenConfig) error {
		tc.Audience = audience
		return nil
	}
}

func WithScope(scope ...string) TokenOption {
	return func(tc *TokenConfig) error {
		tc.Scope = scope
		return nil
	}
}

func WithSubject(subject string) TokenOption {
	return func(tc *TokenConfig) error {
		tc.Subject = subject
		return nil
	}
}

func WithExpiresAt(expiresAt time.Time) TokenOption {
	return func(tc *TokenConfig) error {
		tc.ExpiresAt = expiresAt
		return nil
	}
}

func WithIssuedAt(issuedAt time.Time) TokenOption {
	return func(tc *TokenConfig) error {
		tc.IssuedAt = issuedAt
		return nil
	}
}

// Refresh token options
func WithRefreshAudience(audience ...string) RefreshOption {
	return func(tc *TokenConfig) error {
		tc.Audience = audience
		return nil
	}
}

func WithRefreshScope(scope ...string) RefreshOption {
	return func(tc *TokenConfig) error {
		tc.Scope = scope
		return nil
	}
}

func WithRefreshSubject(subject string) RefreshOption {
	return func(tc *TokenConfig) error {
		tc.Subject = subject
		return nil
	}
}

func WithRefreshExpiresAt(expiresAt time.Time) RefreshOption {
	return func(tc *TokenConfig) error {
		tc.ExpiresAt = expiresAt
		return nil
	}
}

// Token pair options
func WithRefreshToken() TokenPairOption {
	return func(tpc *TokenPairConfig) error {
		tpc.IncludeRefresh = true
		return nil
	}
}

func WithAccessTokenOptions(opts ...TokenOption) TokenPairOption {
	return func(tpc *TokenPairConfig) error {
		tpc.AccessTokenOptions = opts
		return nil
	}
}

func WithRefreshTokenOptions(opts ...RefreshOption) TokenPairOption {
	return func(tpc *TokenPairConfig) error {
		tpc.RefreshTokenOptions = opts
		return nil
	}
}

// Validation options
func SkipExpiration() ValidationOption {
	return func(vc *ValidationConfig) error {
		vc.SkipExpiration = true
		return nil
	}
}

func SkipValidation() ValidationOption {
	return func(vc *ValidationConfig) error {
		vc.SkipValidation = true
		return nil
	}
}

func WithRequiredScopes(scopes ...string) ValidationOption {
	return func(vc *ValidationConfig) error {
		vc.RequiredScopes = scopes
		return nil
	}
}

func WithRequiredIssuer(issuer string) ValidationOption {
	return func(vc *ValidationConfig) error {
		vc.RequiredIssuer = issuer
		return nil
	}
}

func WithRequiredAudience(audience string) ValidationOption {
	return func(vc *ValidationConfig) error {
		vc.RequiredAudience = audience
		return nil
	}
}
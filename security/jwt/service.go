package jwt

type jwtService struct {
	config       Config
	generator    TokenGenerator
	validator    TokenValidator
	refresher    TokenRefresher
	claimsParser ClaimsParser
}

// NewJWTService creates a new JWT service instance
func NewJWTService(config Config) Service {
	parser := NewClaimsParser(config)
	generator := NewTokenGenerator(config, parser)
	validator := NewTokenValidator(config, parser)
	refresher := NewTokenRefresher(validator, generator)

	return &jwtService{
		config:       config,
		generator:    generator,
		validator:    validator,
		refresher:    refresher,
		claimsParser: parser,
	}
}

// Delegation methods for TokenGenerator interface
func (s *jwtService) GenerateAccessToken(claims Claims, opts ...TokenOption) (*SignedToken, error) {
	return s.generator.GenerateAccessToken(claims, opts...)
}

func (s *jwtService) GenerateRefreshToken(claims Claims, opts ...RefreshOption) (*SignedToken, error) {
	return s.generator.GenerateRefreshToken(claims, opts...)
}

func (s *jwtService) GenerateTokenPair(claims Claims, opts ...TokenPairOption) (*TokenPair, error) {
	return s.generator.GenerateTokenPair(claims, opts...)
}

// Delegation methods for TokenValidator interface
func (s *jwtService) ValidateToken(tokenString string, opts ...ValidationOption) (*ValidatedToken, error) {
	return s.validator.ValidateToken(tokenString, opts...)
}

func (s *jwtService) ValidateClaims(claims *Claims, config *ValidationConfig) error {
	return s.validator.ValidateClaims(claims, config)
}

// Delegation methods for TokenRefresher interface
func (s *jwtService) RefreshAccessToken(refreshToken string, opts ...RefreshOption) (*SignedToken, error) {
	return s.refresher.RefreshAccessToken(refreshToken, opts...)
}

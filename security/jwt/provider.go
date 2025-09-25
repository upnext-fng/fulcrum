package jwt

// ServiceProvider NewJWTService creates a new JWT service instance
func ServiceProvider(config Config) Service {
	parser := ParserServiceProvider(config)
	generator := TokenGeneratorProvider(config, parser)
	validator := TokenValidatorProvider(config, parser)
	refresher := TokenRefresherProvider(validator, generator)

	return &jwtService{
		config:       config,
		generator:    generator,
		validator:    validator,
		refresher:    refresher,
		claimsParser: parser,
	}
}

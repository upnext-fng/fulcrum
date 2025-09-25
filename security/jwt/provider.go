package jwt

func NewService(config Config) Service {
	return NewManager(config)
}
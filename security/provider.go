package security

func NewSecurityService(config Config) SecurityService {
	return NewManager(config)
}

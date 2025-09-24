package configuration

func NewConfigurationService(config Config) ConfigurationService {
	return NewManager(config)
}

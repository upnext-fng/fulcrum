package configuration

type ConfigurationService interface {
	LoadConfig(target interface{}) error
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}

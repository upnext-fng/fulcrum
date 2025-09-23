package configuration

import (
	"strings"

	"github.com/spf13/viper"
)

type manager struct {
	viper  *viper.Viper
	config Config
}

func NewManager(config Config) ConfigurationService {
	v := viper.New()

	// Set defaults
	if config.ConfigPath == "" {
		config.ConfigPath = "."
	}
	if config.EnvPrefix == "" {
		config.EnvPrefix = "APP"
	}

	// Configure viper
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(config.ConfigPath)
	v.SetEnvPrefix(config.EnvPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Try to read config file (ignore error if file doesn't exist)
	v.ReadInConfig()

	return &manager{
		viper:  v,
		config: config,
	}
}

func (m *manager) LoadConfig(target interface{}) error {
	return m.viper.Unmarshal(target)
}

func (m *manager) GetString(key string) string {
	return m.viper.GetString(key)
}

func (m *manager) GetInt(key string) int {
	return m.viper.GetInt(key)
}

func (m *manager) GetBool(key string) bool {
	return m.viper.GetBool(key)
}

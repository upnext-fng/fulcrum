package configuration

type Config struct {
	ConfigPath string `mapstructure:"config_path"`
	EnvPrefix  string `mapstructure:"env_prefix"`
}

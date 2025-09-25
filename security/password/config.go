package password

type Config struct {
	Cost int `mapstructure:"hash_cost"`
}
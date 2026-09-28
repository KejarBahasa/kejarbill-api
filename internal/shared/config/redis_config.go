package config

type RedisConfig struct {
	Addr     string `mapstructure:"REDIS_ADDR"`
	Password string `mapstructure:"REDIS_PASSWORD"`
}

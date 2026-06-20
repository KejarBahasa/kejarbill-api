package config

type DBConfig struct {
	Host         string `mapstructure:"DB_HOST"`
	Port         string `mapstructure:"DB_PORT"`
	User         string `mapstructure:"DB_USER"`
	Pass         string `mapstructure:"DB_PASS"`
	Name         string `mapstructure:"DB_NAME"`
	PoolMaxConns int32  `mapstructure:"DB_POOL_MAX_CONNS"`
	PoolMinConns int32  `mapstructure:"DB_POOL_MIN_CONNS"`
}

package config

import (
	"log"
	"time"
	_ "time/tzdata"

	"github.com/spf13/viper"
)

type Config struct {
	AppName string `mapstructure:"APP_NAME"`
	AppPort string `mapstructure:"APP_PORT"`
	AppEnv  string `mapstructure:"APP_ENV"`
	AppTZ   string `mapstructure:"APP_TZ"` // Timezone

	DBHost         string `mapstructure:"DB_HOST"`
	DBPort         string `mapstructure:"DB_PORT"`
	DBUser         string `mapstructure:"DB_USER"`
	DBPass         string `mapstructure:"DB_PASS"`
	DBName         string `mapstructure:"DB_NAME"`
	DBPoolMaxConns int32  `mapstructure:"DB_POOL_MAX_CONNS"`
	DBPoolMinConns int32  `mapstructure:"DB_POOL_MIN_CONNS"`

	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`

	PasetoSecretKey string `mapstructure:"PASETO_SECRET_KEY"`

	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	if config.PasetoSecretKey == "" {
		log.Fatal("PASETO_SECRET_KEY is required")
	}

	if config.AppTZ == "" {
		config.AppTZ = "Asia/Jakarta"
	}

	tz, err := time.LoadLocation(config.AppTZ)
	if err != nil {
		log.Fatalf("Failed to load timezone: %v", err)
	}
	time.Local = tz

	return &config
}

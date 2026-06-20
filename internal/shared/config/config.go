package config

import (
	"log"
	"reflect"
	"time"
	_ "time/tzdata"

	"github.com/spf13/viper"
)

type Config struct {
	App    AppConfig    `mapstructure:",squash"`
	DB     DBConfig     `mapstructure:",squash"`
	Redis  RedisConfig  `mapstructure:",squash"`
	Paseto PasetoConfig `mapstructure:",squash"`
	Auth   AuthConfig   `mapstructure:",squash"`
	Crypto CryptoConfig `mapstructure:",squash"`
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	bindEnvs(reflect.TypeOf(Config{}))

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: env file not found (%v). Using OS environment variables.", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	if config.Paseto.SecretKey == "" {
		log.Fatal("PASETO_SECRET_KEY is required")
	}

	if config.App.TZ == "" {
		config.App.TZ = "Asia/Jakarta"
	}

	tz, err := time.LoadLocation(config.App.TZ)
	if err != nil {
		log.Fatalf("Failed to load timezone: %v", err)
	}
	time.Local = tz

	return &config
}

func bindEnvs(t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tag := f.Tag.Get("mapstructure")

		if f.Type.Kind() == reflect.Struct && tag == ",squash" {
			bindEnvs(f.Type)
			continue
		}

		if tag != "" && tag != ",squash" {
			_ = viper.BindEnv(tag)
		}
	}
}

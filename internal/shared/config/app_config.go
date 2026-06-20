package config

type AppConfig struct {
	Name string `mapstructure:"APP_NAME"`
	Port string `mapstructure:"APP_PORT"`
	Env  string `mapstructure:"APP_ENV"`
	TZ   string `mapstructure:"APP_TZ"`
}

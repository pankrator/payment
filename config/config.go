package config

import (
	"fmt"
	"strings"

	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/web"
	"github.com/spf13/viper"
)

type Settings struct {
	Storage *storage.Settings `mapstructure:"storage"`
	Server  *web.Settings     `mapstructure:"server"`
}

type KeyableSetting interface {
	Keys() []string
}

func Load(config *Config) *Settings {
	settings := &Settings{
		Storage: storage.DefaultSettings(),
		Server:  web.DefaultSettings(),
	}
	all := map[string]KeyableSetting{
		"storage": storage.DefaultSettings(),
		"server":  web.DefaultSettings(),
	}
	for settingName, setting := range all {
		for _, k := range setting.Keys() {
			if err := config.BindEnv(settingName + "." + k); err != nil {
				panic(err)
			}
		}
	}

	if err := config.Unmarshal(settings); err != nil {
		panic(fmt.Errorf("could not unmarshal confi: %s", err))
	}

	return settings
}

type Config struct {
	*viper.Viper
}

func New() *Config {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetEnvPrefix("PAY")

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("could not read config file: %s", err))
	}

	return &Config{
		Viper: v,
	}
}

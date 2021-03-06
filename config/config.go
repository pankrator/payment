package config

import (
	"fmt"
	"strings"

	"github.com/pankrator/payment/services"

	"github.com/pankrator/payment/auth"
	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/users"
	"github.com/pankrator/payment/web"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

type Settings struct {
	Storage *storage.Settings         `mapstructure:"storage"`
	Server  *web.Settings             `mapstructure:"server"`
	Auth    *auth.Settings            `mapstructure:"auth"`
	Users   *users.Settings           `mapstructure:"users"`
	Cleaner *services.CleanerSettings `mapstructure:"cleaner"`
}

type KeyableSetting interface {
	Keys() []string
}

func (s *Settings) Keys() []string {
	keys := make([]string, 0)
	for _, k := range s.Storage.Keys() {
		keys = append(keys, "storage."+k)
	}

	for _, k := range s.Server.Keys() {
		keys = append(keys, "server."+k)
	}

	for _, k := range s.Auth.Keys() {
		keys = append(keys, "auth."+k)
	}

	for _, k := range s.Users.Keys() {
		keys = append(keys, "users."+k)
	}

	for _, k := range s.Cleaner.Keys() {
		keys = append(keys, "cleaner."+k)
	}

	return keys
}

func Load(config *Config) *Settings {
	settings := &Settings{
		Storage: storage.DefaultSettings(),
		Server:  web.DefaultSettings(),
		Auth:    auth.DefaultSettings(),
	}

	if err := config.Unmarshal(settings); err != nil {
		panic(fmt.Errorf("could not unmarshal confi: %s", err))
	}

	return settings
}

type Config struct {
	*viper.Viper
}

func New(configPath string, fs afero.Fs) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath(configPath)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()
	v.SetFs(fs)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("could not read config file: %s", err)
	}

	return &Config{
		Viper: v,
	}, nil
}

func (c *Config) Unmarshal(value KeyableSetting) error {
	for _, k := range value.Keys() {
		if err := c.BindEnv(k); err != nil {
			return err
		}
	}
	return c.Viper.Unmarshal(value)
}

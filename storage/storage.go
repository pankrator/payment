package storage

import (
	"database/sql"
	"errors"

	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/query"
)

var ErrNotFound error = errors.New("not found in storage")

//go:generate counterfeiter . Storage
type Storage interface {
	Open(func(string, string) (*sql.DB, error)) error
	Close()

	Create(object model.Object) (model.Object, error)
	Save(object model.Object) error
	DeleteAll(typee string) error
	Delete(typee string, condition string, args ...interface{}) error
	Get(typee string, id string) (model.Object, error)
	GetBy(typee string, condition string, args ...interface{}) (model.Object, error)
	List(typee string, q ...query.Query) ([]model.Object, error)
	Count(typee string, condition string, args ...interface{}) (int, error)

	Transaction(f func(s Storage) error) error
}

type Settings struct {
	Host              string `mapstructure:"host"`
	Port              string `mapstructure:"port"`
	Database          string `mapstructure:"database"`
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	SkipSSLValidation bool   `mapstructure:"skip_ssl_validation"`
}

func DefaultSettings() *Settings {
	return &Settings{
		Host:              "127.0.0.1",
		Port:              "5432",
		Database:          "payment",
		Username:          "payment",
		Password:          "payment",
		SkipSSLValidation: true,
	}
}

func (s *Settings) Keys() []string {
	return []string{
		"host",
		"port",
		"database",
		"username",
		"password",
		"skip_ssl_validation",
	}
}

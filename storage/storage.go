package storage

import (
	"errors"

	"github.com/pankrator/payment/model"
)

var ErrNotFound error = errors.New("not found in storage")

type Storage interface {
	Open() error
	Close()
	RegisterModels(typee string, modelProvider func() Model)

	Create(object model.Object) (model.Object, error)
	Save(object model.Object) error
	Get(typee string, id string) (model.Object, error)
	Count(typee string, condition string, args ...interface{}) (int, error)

	Transaction(f func(s Storage) error) error
}

type Settings struct {
	Host              string
	Port              string
	Database          string
	Username          string
	Password          string
	SkipSSLValidation bool
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

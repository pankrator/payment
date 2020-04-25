package gormdb

import (
	"github.com/jinzhu/gorm"
	"github.com/pankrator/payment/model"
)

type Model interface {
	InitSQL(*gorm.DB) error
	FromObject(o model.Object) (Model, error)
	ToObject() model.Object
}

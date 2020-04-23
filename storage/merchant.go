package storage

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/pankrator/payment/model"
)

type Merchant struct {
	*gorm.Model
	Name                string
	Description         string
	Email               string
	Status              bool
	TotalTransactionSum int64
}

func (m *Merchant) InitSQL(*gorm.DB) error {
	return nil
}

func (m *Merchant) ToObject() model.Object {
	return &model.Merchant{}
}

func (m *Merchant) FromObject(o model.Object) Model {
	_, ok := o.(*model.Merchant)
	if !ok {
		panic(fmt.Sprintf("%s is not merchant", o.GetType()))
	}
	return &Merchant{}
}

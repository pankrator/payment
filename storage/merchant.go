package storage

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pankrator/payment/model"
)

type Merchant struct {
	UUID                string `gorm:"primary_key"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Name                string `gorm:"type:varchar(300);unique;not null"`
	Description         string
	Email               string `gorm:"type:varchar(300);unique;not null"`
	Status              bool
	TotalTransactionSum int64
}

func (m *Merchant) InitSQL(*gorm.DB) error {
	return nil
}

func (m *Merchant) ToObject() model.Object {
	return &model.Merchant{
		UUID:                m.UUID,
		Name:                m.Name,
		Description:         m.Description,
		Email:               m.Email,
		Status:              m.Status,
		TotalTransactionSum: m.TotalTransactionSum,
	}
}

func (m *Merchant) FromObject(o model.Object) Model {
	merchant, ok := o.(*model.Merchant)
	if !ok {
		panic(fmt.Sprintf("%s is not merchant", o.GetType()))
	}
	return &Merchant{
		UUID:                merchant.UUID,
		Name:                merchant.Name,
		Description:         merchant.Description,
		Email:               merchant.Email,
		Status:              merchant.Status,
		TotalTransactionSum: merchant.TotalTransactionSum,
	}
}

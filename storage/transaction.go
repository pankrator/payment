package storage

import (
	"database/sql/driver"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/pankrator/payment/model"
)

type TransactionState string

const (
	Approved TransactionState = "approved"
	Reversed TransactionState = "reversed"
	Refunded TransactionState = "refunded"
	Errored  TransactionState = "error"
)

func (s *TransactionState) Scan(value interface{}) error {
	*s = TransactionState(value.([]byte))
	return nil
}

func (p TransactionState) Value() (driver.Value, error) {
	return string(p), nil
}

type TransactionType string

const (
	Authorize TransactionType = "authorize"
	Charge    TransactionType = "charge"
	Refund    TransactionType = "refund"
	Reversal  TransactionType = "reversal"
)

func (s *TransactionType) Scan(value interface{}) error {
	*s = TransactionType(value.([]byte))
	return nil
}

func (p TransactionType) Value() (driver.Value, error) {
	return string(p), nil
}

type Transaction struct {
	*gorm.Model
	UUID          *string         `gorm:"type:varchar(100);unique;not null"`
	Type          TransactionType `gorm:"type:transaction_type"`
	Amount        int
	CustomerEmail string
	CustomerPhone string
	Status        TransactionState `gorm:"type:transaction_status"`

	Merchant   *Merchant `gorm:"foreignKey:MerchantID"`
	MerchantID *int

	DependsOn     *Transaction `gorm:"foreignkey:TransactionID"`
	TransactionID *int
}

func (t *Transaction) InitSQL(db *gorm.DB) error {
	err := db.Model(t).
		AddForeignKey("transaction_id", "transactions(id)", "RESTRICT", "RESTRICT").
		AddForeignKey("merchant_id", "merchants(id)", "RESTRICT", "RESTRICT").
		AddUniqueIndex("unique_uuid", "uuid").
		Error

	return err
}

func (t *Transaction) ToObject() model.Object {
	return &model.Transaction{
		UUID:          *t.UUID,
		Amount:        t.Amount,
		CustomerEmail: t.CustomerEmail,
		CustomerPhone: t.CustomerPhone,
		Type:          model.TransactionType(t.Type),
		Status:        model.TransactionState(t.Status),
	}
}

func (t *Transaction) FromObject(o model.Object) Model {
	transaction, ok := o.(*model.Transaction)
	if !ok {
		panic(fmt.Sprintf("%s is not transaction", o.GetType()))
	}
	return &Transaction{
		UUID:          &transaction.UUID,
		Amount:        transaction.Amount,
		CustomerEmail: transaction.CustomerEmail,
		CustomerPhone: transaction.CustomerPhone,
		Type:          TransactionType(transaction.Type),
		Status:        TransactionState(transaction.Status),
	}
}

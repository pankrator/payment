package storage

import (
	"database/sql/driver"
	"fmt"
	"time"

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
	UUID      string `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Type          TransactionType `gorm:"type:transaction_type"`
	Amount        int
	CustomerEmail string
	CustomerPhone string
	Status        TransactionState `gorm:"type:transaction_status"`

	Merchant   *Merchant `gorm:"foreignKey:MerchantID"`
	MerchantID string

	DependsOn     *Transaction `gorm:"foreignkey:TransactionID"`
	TransactionID *string
}

func (t *Transaction) InitSQL(db *gorm.DB) error {
	err := db.Model(t).
		AddForeignKey("transaction_id", "transactions(uuid)", "RESTRICT", "RESTRICT").
		AddForeignKey("merchant_id", "merchants(uuid)", "RESTRICT", "RESTRICT").
		Error

	return err
}

func (t *Transaction) ToObject() model.Object {
	result := &model.Transaction{
		UUID:          t.UUID,
		Amount:        t.Amount,
		CustomerEmail: t.CustomerEmail,
		CustomerPhone: t.CustomerPhone,
		Type:          model.TransactionType(t.Type),
		Status:        model.TransactionState(t.Status),
		MerchantID:    t.MerchantID,
	}

	if t.TransactionID != nil {
		result.DependsOnUUID = *t.TransactionID
	}

	return result
}

func (t *Transaction) FromObject(o model.Object) (Model, error) {
	transaction, ok := o.(*model.Transaction)
	if !ok {
		return nil, fmt.Errorf("%s is not transaction", o.GetType())
	}
	result := &Transaction{
		UUID:          transaction.UUID,
		MerchantID:    transaction.MerchantID,
		Amount:        transaction.Amount,
		CustomerEmail: transaction.CustomerEmail,
		CustomerPhone: transaction.CustomerPhone,
		Type:          TransactionType(transaction.Type),
		Status:        TransactionState(transaction.Status),
	}

	if transaction.DependsOnUUID != "" {
		result.TransactionID = &transaction.DependsOnUUID
	}
	return result, nil
}

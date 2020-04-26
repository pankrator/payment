package model

import (
	"encoding/xml"
	"errors"
	"fmt"
)

const TransactionObjectType string = "Transaction"

type TransactionState string

const (
	Approved TransactionState = "approved"
	Reversed TransactionState = "reversed"
	Refunded TransactionState = "refunded"
	Errored  TransactionState = "errored"
)

type TransactionType string

const (
	Authorize TransactionType = "authorize"
	Charge    TransactionType = "charge"
	Refund    TransactionType = "refund"
	Reversal  TransactionType = "reversal"
)

type Transaction struct {
	XMLName       xml.Name         `xml:"Transaction" json:"-"`
	UUID          string           `json:"uuid" xml:"UUID"`
	Type          TransactionType  `json:"type" xml:"Type"`
	Amount        int              `json:"amount" xml:"Amount"`
	CustomerEmail string           `json:"customer_email" xml:"CustomerEmail"`
	CustomerPhone string           `json:"customer_phone" xml:"CustomerPhone"`
	Status        TransactionState `json:"status" xml:"Status"`

	DependsOnUUID string `json:"depends_on_uuid" xml:"DependsOnUUID"`
	MerchantID    string `json:"merchant_id" xml:"MerchantID"`
}

func (t *Transaction) GetType() string {
	return TransactionObjectType
}

func (t *Transaction) Validate() error {
	if t.MerchantID == "" {
		return errors.New("merchant id is required for transaction")
	}
	if t.Amount < 1 && t.Type != Reversal {
		return errors.New("amount should be greater than 0")
	}
	if t.Status != "" {
		return errors.New("status should not be provided")
	}
	switch t.Type {
	case Authorize:
		if t.DependsOnUUID != "" {
			return fmt.Errorf("transaction of type %s cannot depend on another transaction", Authorize)
		}
	default:
		if t.DependsOnUUID == "" {
			return fmt.Errorf("transaction of type %s should depend on another transaction", t.Type)
		}
	}
	return nil
}

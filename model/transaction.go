package model

import "encoding/xml"

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
	XMLName       xml.Name `xml:"Transaction" json:"-"`
	UUID          string
	Type          TransactionType  `json:"type" xml:"Type"`
	Amount        int              `json:"amount" xml:"Amount"`
	CustomerEmail string           `json:"customer_email" xml:"CustomerEmail"`
	CustomerPhone string           `json:"customer_phone" xml:"CustomerPhone"`
	Status        TransactionState `json:"status" xml:"Status"`
}

func (t *Transaction) GetType() string {
	return TransactionObjectType
}

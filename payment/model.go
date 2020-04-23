package payment

import "encoding/xml"

type TransactionState string

const (
	Approved TransactionState = "approved"
	Reversed TransactionState = "reversed"
	Refunded TransactionState = "refunded"
	Errored  TransactionState = "errored"
)

type Transaction struct {
	XMLName       xml.Name `xml:"Transaction"`
	UUID          string
	Amount        int              `json:"amount" xml:"Amount"`
	CustomerEmail string           `json:"customer_email" xml:"CustomerEmail"`
	CustomerPhone string           `json:"customer_phone" xml:"CustomerPhone"`
	Status        TransactionState `json:"state" xml:"Status"`
}

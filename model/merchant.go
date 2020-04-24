package model

const MerchantType string = "Merchant"

type Merchant struct {
	UUID                string
	Name                string
	Description         string
	Email               string
	Status              bool
	TotalTransactionSum int64
}

func (m *Merchant) GetType() string {
	return MerchantType
}

func (m *Merchant) Validate() error {
	return nil
}

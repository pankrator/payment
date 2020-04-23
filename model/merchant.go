package model

const MerchantType string = "Merchant"

type Merchant struct {
}

func (m *Merchant) GetType() string {
	return MerchantType
}

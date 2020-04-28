package model

import "github.com/pankrator/payment/users"

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

func MerchantFromUser(user users.User) *Merchant {
	return &Merchant{
		Name:                user.Name,
		Email:               user.Email,
		Description:         user.Description,
		Status:              user.Status == "active",
		TotalTransactionSum: 0,
	}
}

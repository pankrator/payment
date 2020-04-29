package model

import "github.com/pankrator/payment/users"

const MerchantType string = "Merchant"

type Merchant struct {
	UUID                string `json:"uuid"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	Email               string `json:"email"`
	Status              bool   `json:"status"`
	TotalTransactionSum int64  `json:"total_transaction_sum"`
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

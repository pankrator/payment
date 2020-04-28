package app

import "github.com/pankrator/payment/model"

type MerchantService interface {
	Create(*model.Merchant) (model.Object, error)
}

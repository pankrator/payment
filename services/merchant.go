package services

import (
	"errors"
	"log"

	"github.com/gofrs/uuid"
	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/query"
	"github.com/pankrator/payment/storage"
)

type MerchantService struct {
	repository storage.Storage
}

func NewMerchantService(repository storage.Storage) *MerchantService {
	return &MerchantService{
		repository: repository,
	}
}

func (ms *MerchantService) Create(merchant *model.Merchant) (model.Object, error) {
	if err := merchant.Validate(); err != nil {
		return nil, err
	}

	UUID, err := uuid.NewV4()
	if err != nil {
		log.Printf("Could not generate UUID: %s", err)
		return nil, errors.New("could not generate UUID")
	}

	merchant.UUID = UUID.String()

	return ms.repository.Create(merchant)
}

func (ms *MerchantService) Get(uuid string) (*model.Merchant, error) {
	object, err := ms.repository.Get(model.MerchantType, uuid)
	if err != nil {
		return nil, err
	}
	return object.(*model.Merchant), nil
}

func (ms *MerchantService) List(q []query.Query) ([]model.Object, error) {
	return ms.repository.List(model.MerchantType, q...)
}

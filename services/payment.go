package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/storage"
)

type PaymentService struct {
	repository storage.Storage
}

func NewPaymentService(repository storage.Storage) *PaymentService {
	return &PaymentService{
		repository: repository,
	}
}

func (ps *PaymentService) Create(transaction *model.Transaction) (model.Object, error) {
	if err := transaction.Validate(); err != nil {
		return nil, err
	}

	UUID, err := uuid.NewV4()
	if err != nil {
		log.Printf("Could not generate UUID: %s", err)
		return nil, errors.New("could not generate UUID")
	}

	transaction.UUID = UUID.String()
	transaction.Status = model.Approved

	if err := ps.checkMerchantStatus(transaction); err != nil {
		return nil, err
	}

	var parentTransaction *model.Transaction

	// TODO: Do everything in one transaction, otherwise it might happen that two requests change dependent entities or so and
	// might get into inconsistent state
	if transaction.Type != model.Authorize {
		count, err := ps.repository.Count(model.TransactionObjectType, "transaction_id = ?", transaction.DependsOnUUID)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("the parent transaction is already followed")
		}
		parentTransaction, err = ps.findParentTransaction(ps.repository, transaction)
		if err != nil {
			return nil, err
		}

		if err = ps.checkParentTransactionConditions(transaction, parentTransaction); err != nil {
			return nil, err
		}

		ps.updateStateBasedOnParent(transaction, parentTransaction)
	}

	switch transaction.Type {
	case model.Authorize:
		return ps.repository.Create(transaction)
	case model.Charge:
		return ps.chargeTransaction(transaction)
	case model.Refund:
		return ps.refundTransaction(transaction, parentTransaction)
	case model.Reversal:
		return ps.reverseTransaction(transaction, parentTransaction)
	default:
		return nil, fmt.Errorf("transaction type %s not recognized", transaction.Type)
	}
}

func (ps *PaymentService) chargeTransaction(transaction *model.Transaction) (model.Object, error) {
	var result model.Object
	err := ps.repository.Transaction(func(tx storage.Storage) error {
		var err error
		result, err = tx.Create(transaction)
		if err != nil {
			return fmt.Errorf("database operation failed: %s", err)
		}

		if transaction.Status != model.Errored {
			object, err := tx.Get(model.MerchantType, transaction.MerchantID)
			if err != nil {
				return err
			}
			merchant := object.(*model.Merchant)
			merchant.TotalTransactionSum += int64(transaction.Amount)
			return tx.Save(merchant)
		}
		return nil
	})

	return result, err
}

func (ps *PaymentService) refundTransaction(transaction, parentTransaction *model.Transaction) (model.Object, error) {
	var result model.Object
	err := ps.repository.Transaction(func(tx storage.Storage) error {
		var err error
		result, err = tx.Create(transaction)
		if err != nil {
			return fmt.Errorf("database operation failed: %s", err)
		}

		if transaction.Status != model.Errored {
			parentTransaction.Status = model.Refunded
			if err := tx.Save(parentTransaction); err != nil {
				return err
			}

			object, err := tx.Get(model.MerchantType, transaction.MerchantID)
			if err != nil {
				return err
			}

			merchant := object.(*model.Merchant)
			merchant.TotalTransactionSum -= int64(transaction.Amount)
			return tx.Save(merchant)
		}

		return nil
	})
	return result, err
}

func (ps *PaymentService) reverseTransaction(transaction, parentTransaction *model.Transaction) (model.Object, error) {
	var result model.Object
	err := ps.repository.Transaction(func(tx storage.Storage) error {
		var err error
		result, err = tx.Create(transaction)
		if err != nil {
			return fmt.Errorf("database operation failed: %s", err)
		}

		parentTransaction.Status = model.Reversed
		return tx.Save(parentTransaction)
	})
	return result, err
}

func (ps *PaymentService) checkParentTransactionConditions(transaction *model.Transaction, parent *model.Transaction) error {
	switch transaction.Type {
	case model.Reversal:
		fallthrough
	case model.Charge:
		if parent.Type != model.Authorize {
			return fmt.Errorf("parent transaction should be of type %s", model.Authorize)
		}
	case model.Refund:
		if parent.Type != model.Charge {
			return fmt.Errorf("parent transaction should be of type %s", model.Charge)
		}
	}
	return nil
}

func (ps *PaymentService) updateStateBasedOnParent(transaction *model.Transaction, parent *model.Transaction) {
	switch transaction.Type {
	case model.Reversal:
		fallthrough
	case model.Charge:
		if parent.Status != model.Approved {
			transaction.Status = model.Errored
		}
	case model.Refund:
		if parent.Status != model.Approved {
			transaction.Status = model.Errored
		}
	}
}

func (ps *PaymentService) checkMerchantStatus(transaction *model.Transaction) error {
	object, err := ps.repository.Get(model.MerchantType, transaction.MerchantID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fmt.Errorf("merchant with id %s not found", transaction.MerchantID)
		}
		return err
	}
	merchant := object.(*model.Merchant)
	if !merchant.Status {
		return fmt.Errorf("merchant with name %s is not active", merchant.Name)
	}
	return nil
}

func (ps *PaymentService) findParentTransaction(repository storage.Storage, transaction *model.Transaction) (*model.Transaction, error) {
	object, err := repository.Get(model.TransactionObjectType, transaction.DependsOnUUID)
	if err != nil {
		if err == storage.ErrNotFound {
			return nil, fmt.Errorf("parent transaction with uuid %s not found", transaction.DependsOnUUID)
		}
		return nil, err
	}
	return object.(*model.Transaction), nil
}

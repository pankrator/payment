package services

import (
	"fmt"

	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/storage"
)

type PaymentService struct {
	Repository storage.Storage
}

func NewPaymentService(repository storage.Storage) *PaymentService {
	return &PaymentService{
		Repository: repository,
	}
}

func (ps *PaymentService) Create(object model.Object) (model.Object, error) {
	transaction := object.(*model.Transaction)
	transaction.Status = model.Approved

	if err := ps.checkMerchantStatus(transaction); err != nil {
		return nil, err
	}

	var parentTransaction *model.Transaction

	if transaction.Type != model.Authorize {
		count, err := ps.Repository.Count(model.TransactionObjectType, "transaction_id = ?", transaction.DependsOnUUID)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("this transaction is already followed")
		}
		parentTransaction, err = ps.findParentTransaction(ps.Repository, transaction)
		if err != nil {
			return nil, err
		}

		if err = ps.checkParentTransactionConditions(transaction, parentTransaction); err != nil {
			return nil, err
		}
	}

	switch transaction.Type {
	case model.Authorize:
		return ps.Repository.Create(transaction)
	case model.Charge:
		return ps.chargeTransaction(transaction)
	case model.Refund:
		return ps.refundTransaction(transaction, parentTransaction)
	default:
		return nil, fmt.Errorf("transaction type %s not recognized", transaction.Type)
	}
}

func (ps *PaymentService) chargeTransaction(transaction *model.Transaction) (model.Object, error) {
	var result model.Object
	err := ps.Repository.Transaction(func(tx storage.Storage) error {
		var err error
		result, err = tx.Create(transaction)
		if err != nil {
			return fmt.Errorf("database operation failed: %s", err)
		}
		object, err := tx.Get(model.MerchantType, transaction.MerchantID)
		if err != nil {
			return err
		}
		merchant := object.(*model.Merchant)
		merchant.TotalTransactionSum += int64(transaction.Amount)
		return tx.Save(merchant)
	})

	return result, err
}

func (ps *PaymentService) refundTransaction(transaction, parentTransaction *model.Transaction) (model.Object, error) {
	var result model.Object
	err := ps.Repository.Transaction(func(tx storage.Storage) error {
		var err error
		result, err = tx.Create(transaction)
		if err != nil {
			return fmt.Errorf("database operation failed: %s", err)
		}

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
	})
	return result, err
}

func (ps *PaymentService) checkParentTransactionConditions(transaction *model.Transaction, parent *model.Transaction) error {
	switch transaction.Type {
	case model.Charge:
		if parent.Type != model.Authorize {
			return fmt.Errorf("parent transaction should be of type %s", model.Authorize)
		}
		if parent.Status != model.Approved {
			return fmt.Errorf("authorize transaction should be approved, but is %s", parent.Status)
		}
		// TODO: Check if parent transaction is not already referenced by another transaction
	case model.Refund:
		if parent.Type != model.Charge {
			return fmt.Errorf("parent transaction should be of type %s", model.Charge)
		}
		if parent.Status != model.Approved {
			return fmt.Errorf("cannot refund charge transaction that is in state %s", parent.Status)
		}
	}
	return nil
}

func (ps *PaymentService) checkMerchantStatus(transaction *model.Transaction) error {
	object, err := ps.Repository.Get(model.MerchantType, transaction.MerchantID)
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

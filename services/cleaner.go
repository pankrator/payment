package services

import (
	"context"
	"log"
	"time"

	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/storage"
)

type CleanerSettings struct {
	KeepTransactionsFor time.Duration `mapstructure:"keep_transactions_for"`
	Interval            time.Duration `mapstructure:"interval"`
}

func (s *CleanerSettings) Keys() []string {
	return []string{
		"keep_transactions_for",
		"interval",
	}
}

type TransactionClenaer struct {
	settings   *CleanerSettings
	repository storage.Storage
}

func NewTransactionCleaner(settings *CleanerSettings, repository storage.Storage) *TransactionClenaer {
	return &TransactionClenaer{
		settings:   settings,
		repository: repository,
	}
}

func (tc *TransactionClenaer) Start(ctx context.Context) {
	go func() {
		for {
			elapsed := time.After(tc.settings.Interval)
			select {
			case <-ctx.Done():
				log.Printf("Context cancelled. Stopping the transaction cleaner...")
				return
			case <-elapsed:
				log.Printf("Cleaning old transactions...")
				if err := tc.run(ctx); err != nil {
					log.Printf("Could not delete old transaction: %s", err)
				}
				log.Printf("Finished cleaning old transaction...")
			}
		}
	}()

	log.Printf("Transaction cleaner started")
}

func (tc *TransactionClenaer) run(ctx context.Context) error {
	return tc.repository.Delete(model.TransactionObjectType, "created_at < ?", time.Now().Add(-tc.settings.KeepTransactionsFor))
}

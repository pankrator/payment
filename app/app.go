package app

import (
	"context"

	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/payment"
	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/web"
)

type App struct {
	server  *web.Server
	storage *storage.Storage
}

func New() *App {
	s := storage.New(storage.DefaultSettings())

	api := &web.Api{
		Controllers: []web.Controller{
			&payment.Controller{
				Repository: s,
			},
		},
	}
	server := web.NewServer(web.DefaultSettings(), api)
	return &App{
		server:  server,
		storage: s,
	}
}

func (a *App) Start(ctx context.Context) {
	a.registerModels()

	if err := a.storage.Open(); err != nil {
		panic(err)
	}
	a.server.Run(ctx)
}

func (a *App) registerModels() {
	a.storage.RegisterModels(model.TransactionObjectType, &storage.Transaction{})
	a.storage.RegisterModels(model.MerchantType, &storage.Merchant{})
}

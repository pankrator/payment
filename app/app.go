package app

import (
	"context"

	"github.com/pankrator/payment/api"
	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/services"
	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/web"
)

type App struct {
	server     *web.Server
	repository *storage.Storage
}

func New() *App {
	repository := storage.New(storage.DefaultSettings())

	paymentService := services.NewPaymentService(repository)

	api := &web.Api{
		Controllers: []web.Controller{
			api.NewPaymentController(paymentService),
		},
	}

	server := web.NewServer(web.DefaultSettings(), api)
	return &App{
		server:     server,
		repository: repository,
	}
}

func (a *App) Start(ctx context.Context) {
	a.registerModels()

	if err := a.repository.Open(); err != nil {
		panic(err)
	}
	a.server.Run(ctx)
}

func (a *App) registerModels() {
	a.repository.RegisterModels(model.TransactionObjectType, func() storage.Model { return &storage.Transaction{} })
	a.repository.RegisterModels(model.MerchantType, func() storage.Model { return &storage.Merchant{} })
}

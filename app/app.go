package app

import (
	"context"

	"github.com/pankrator/payment/config"

	"github.com/pankrator/payment/api"
	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/services"
	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/storage/gormdb"
	"github.com/pankrator/payment/web"
)

type App struct {
	server     *web.Server
	repository storage.Storage
}

func New() *App {
	web.RegisterParser("application/xml", &web.XMLParser{})
	web.RegisterParser("application/json", &web.JSONParser{})

	cfg := config.New()
	settings := config.Load(cfg)

	repository := gormdb.New(settings.Storage)

	paymentService := services.NewPaymentService(repository)

	api := &web.Api{
		Controllers: []web.Controller{
			api.NewPaymentController(paymentService),
		},
	}

	server := web.NewServer(settings.Server, api)
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

package app

import (
	"context"
	"database/sql"

	"github.com/pankrator/payment/config"
	"github.com/spf13/afero"

	"github.com/pankrator/payment/api"
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

	cfg, err := config.New(afero.NewOsFs())
	if err != nil {
		panic(err)
	}

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
	if err := a.repository.Open(func(driver, url string) (*sql.DB, error) {
		return sql.Open(driver, url)
	}); err != nil {
		panic(err)
	}
	a.server.Run(ctx)
}

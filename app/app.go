package app

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pankrator/payment/uaa"
	"github.com/pankrator/payment/users"

	"github.com/pankrator/payment/api/auth"

	oauth "github.com/pankrator/payment/auth"
	"github.com/pankrator/payment/config"
	"github.com/spf13/afero"

	"github.com/pankrator/payment/api"
	"github.com/pankrator/payment/services"
	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/storage/gormdb"
	"github.com/pankrator/payment/web"
)

type App struct {
	Server     *web.Server
	Repository storage.Storage
	UaaClient  *uaa.UAAClient
	Settings   *config.Settings
}

func New(configFileLocation string) *App {
	web.RegisterParser("application/xml", &web.XMLParser{})
	web.RegisterParser("application/json", &web.JSONParser{})

	ctx := context.Background()

	cfg, err := config.New(configFileLocation, afero.NewOsFs())
	if err != nil {
		panic(err)
	}
	settings := config.Load(cfg)

	uaaClient, err := uaa.NewClient(&uaa.UAAConfig{
		Auth: &oauth.Config{
			ClientID:          settings.Auth.AdminClientID,
			ClientSecret:      settings.Auth.AdminClientSecret,
			SkipSSLValidation: false,
			Timeout:           time.Second * 10,
		},
		URL: settings.Auth.OauthServerURL,
	})
	if err != nil {
		panic(fmt.Errorf("could not build uaa client: %s", err))
	}

	authenticator, err := auth.NewTokenAuthenticator(ctx, settings.Auth)
	if err != nil {
		panic(fmt.Errorf("could not build authenticator: %s", err))
	}
	authFilter := auth.NewFilter(authenticator)

	repository := gormdb.New(settings.Storage)
	paymentService := services.NewPaymentService(repository)

	api := &web.Api{
		Controllers: []web.Controller{
			api.NewPaymentController(paymentService),
		},
		Filters: []web.Filter{
			authFilter,
		},
	}

	server := web.NewServer(settings.Server, api)
	return &App{
		Server:     server,
		Repository: repository,
		UaaClient:  uaaClient,
		Settings:   settings,
	}
}

func (a *App) InitUsers(ctx context.Context) {
	reader := users.NewReader(a.Settings.Users)
	if err := reader.Load(); err != nil {
		panic(fmt.Errorf("Could not read users: %s", err))
	}
	for _, user := range reader.Users {
		err := a.UaaClient.CreateUser(ctx, user.Name, user.Email, user.Password)
		if err != nil {
			panic(fmt.Errorf("could not create user: %s", err))
		}
	}
}

func (a *App) Start(ctx context.Context) {
	if err := a.Repository.Open(func(driver, url string) (*sql.DB, error) {
		return sql.Open(driver, url)
	}); err != nil {
		panic(err)
	}
	a.Server.Run(ctx)
}

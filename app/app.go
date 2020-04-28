package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/pankrator/payment/auth"
	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/uaa"
	"github.com/pankrator/payment/users"

	"github.com/pankrator/payment/api/filter"

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

	transactionCleaner *services.TransactionClenaer
	merchantService    MerchantService
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
	authFilter := filter.NewAuthFilter(authenticator)

	repository := gormdb.New(settings.Storage)
	paymentService := services.NewPaymentService(repository)
	merchantService := services.NewMerchantService(repository)

	transactionCleaner := services.NewTransactionCleaner(settings.Cleaner, repository)

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
		Server:             server,
		Repository:         repository,
		UaaClient:          uaaClient,
		Settings:           settings,
		merchantService:    merchantService,
		transactionCleaner: transactionCleaner,
	}
}

func (a *App) initUsers(ctx context.Context, usersData []users.User, groupNames []string) {
	groups := make([]*uaa.Group, 0)
	for _, groupName := range groupNames {
		g, err := a.UaaClient.GetGroup(ctx, groupName)
		if err != nil {
			panic(fmt.Errorf("could not get uaa groups: %s", err))
		}
		groups = append(groups, g)
	}

	for _, user := range usersData {
		if user.Type == users.Merchant {
			count, err := a.Repository.Count(model.MerchantType, "email = ?", user.Email)
			if err != nil {
				panic(err)
			}
			if count < 1 {
				_, err = a.merchantService.Create(model.MerchantFromUser(user))
				if err != nil {
					panic(fmt.Errorf("could not create merchant: %s", err))
				}
			} else {
				log.Println("Merchant already created")
			}
		}

		userID, err := a.UaaClient.CreateUser(ctx, user.Name, user.Email, user.Password)
		if err != nil {
			panic(fmt.Errorf("could not create user: %s", err))
		}
		if userID != "" {
			for _, group := range groups {
				if added, err := a.UaaClient.AddUserToGroup(ctx, userID, group); err != nil {
					panic(fmt.Errorf("could not add user to group: %s", err))
				} else if added {
					log.Printf("User %s added to group %s", user.Name, group.DisplayName)
				}
			}
		}
	}
}

func (a *App) loadUsers() *users.UserReader {
	reader := users.NewCSVReader(a.Settings.Users)
	if err := reader.Load(); err != nil {
		panic(fmt.Errorf("Could not read users: %s", err))
	}
	return reader
}

func splitUsersByType(reader *users.UserReader) map[users.UserType][]users.User {
	result := make(map[users.UserType][]users.User)
	for _, user := range reader.Users {
		if _, found := result[user.Type]; !found {
			result[user.Type] = make([]users.User, 0)
		}
		result[user.Type] = append(result[user.Type], user)
	}
	return result
}

func (a *App) Start(ctx context.Context) {
	if err := a.Repository.Open(func(driver, url string) (*sql.DB, error) {
		return sql.Open(driver, url)
	}); err != nil {
		panic(err)
	}

	reader := a.loadUsers()
	usersByType := splitUsersByType(reader)

	adminGroups := []string{"merchant.read", "merchant.write", "merchant.delete", "transaction.write", "transaction.read"}
	merchantGroups := []string{"transaction.write", "transaction.read"}

	a.initUsers(
		ctx,
		usersByType[users.Admin],
		adminGroups,
	)
	a.initUsers(
		ctx,
		usersByType[users.Merchant],
		merchantGroups,
	)

	a.transactionCleaner.Start(ctx)
	a.Server.Run(ctx)
}

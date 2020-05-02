package app

import (
	"context"
	"fmt"
	"log"

	"github.com/pankrator/payment/api"
	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/uaa"
	"github.com/pankrator/payment/users"
)

type UserIniator struct {
	UaaClient       *uaa.UAAClient
	Repository      storage.Storage
	MerchantService api.MerchantService
}

func NewUserInitiator(
	uaaClient *uaa.UAAClient,
	repository storage.Storage,
	merchantService api.MerchantService) *UserIniator {
	return &UserIniator{
		UaaClient:       uaaClient,
		Repository:      repository,
		MerchantService: merchantService,
	}
}

func (ui *UserIniator) LoadUsers(settings *users.Settings) map[users.UserType][]users.User {
	reader := users.NewCSVReader(settings)
	if err := reader.Load(); err != nil {
		panic(fmt.Errorf("Could not read users: %s", err))
	}

	return splitUsersByType(reader)
}

func (ui *UserIniator) InitUsers(ctx context.Context, usersData []users.User, groupNames []string) {
	groups := make([]*uaa.Group, 0)
	for _, groupName := range groupNames {
		g, err := ui.UaaClient.GetGroup(ctx, groupName)
		if err != nil {
			panic(fmt.Errorf("could not get uaa groups: %s", err))
		}
		groups = append(groups, g)
	}

	for _, user := range usersData {
		if user.Type == users.Merchant {
			count, err := ui.Repository.Count(model.MerchantType, "email = ?", user.Email)
			if err != nil {
				panic(err)
			}
			if count < 1 {
				_, err = ui.MerchantService.Create(model.MerchantFromUser(user))
				if err != nil {
					panic(fmt.Errorf("could not create merchant: %s", err))
				}
			} else {
				log.Println("Merchant already created")
			}
		}

		userID, err := ui.UaaClient.CreateUser(ctx, user.Name, user.Email, user.Password)
		if err != nil {
			panic(fmt.Errorf("could not create user: %s", err))
		}
		if userID != "" {
			for _, group := range groups {
				if added, err := ui.UaaClient.AddUserToGroup(ctx, userID, group); err != nil {
					panic(fmt.Errorf("could not add user to group: %s", err))
				} else if added {
					log.Printf("User %s added to group %s", user.Name, group.DisplayName)
				}
			}
		}
	}
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

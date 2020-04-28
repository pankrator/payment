package uaa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pankrator/payment/auth"
	"github.com/pankrator/payment/web"
)

type UAAConfig struct {
	Auth *auth.Config
	URL  string
}

type UAAClient struct {
	client auth.Client
	config *UAAConfig
}

func NewClient(config *UAAConfig) (*UAAClient, error) {
	info, err := auth.GetInfo(config.URL)
	if err != nil {
		return nil, err
	}
	config.Auth.TokenEndpoint = info.TokenEndpoint
	client := auth.New(config.Auth)

	return &UAAClient{
		client: client,
		config: config,
	}, nil
}

type emailValue struct {
	Value string `json:"value"`
}

type createUserRequest struct {
	Emails   []emailValue `json:"emails"`
	UserName string       `json:"userName"`
	Verified bool         `json:"verified"`
	Active   bool         `json:"active"`
	Password string       `json:"password"`
}

func (uc *UAAClient) CreateUser(ctx context.Context, username, email, password string) error {
	createUserBytes, err := json.Marshal(&createUserRequest{
		Emails: []emailValue{
			{
				Value: email,
			},
		},
		UserName: username,
		Password: password,
		Active:   true,
		Verified: true,
	})
	if err != nil {
		return fmt.Errorf("could not marshal create user request: %s", err)
	}
	reader := bytes.NewReader(createUserBytes)

	log.Printf("Requesting UAA to endpoint %s", uc.config.URL+"/Users")
	req, err := http.NewRequest(http.MethodPost, uc.config.URL+"/Users", reader)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}

	resp, err := uc.client.Do(req)
	if err != nil {
		return err
	}
	var result map[string]interface{}
	err = web.BodyToObject(resp.Body, &result)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusConflict:
		log.Printf("User with name %s already exists, skipping it", username)
		return nil
	case http.StatusCreated:
		return nil
	default:
		return fmt.Errorf("uaa responded with status code: %d and error: %s", resp.StatusCode, result["error_description"].(string))
	}
}

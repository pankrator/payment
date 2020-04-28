package auth

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type OAuthClient struct {
	*http.Client
}

type Config struct {
	Timeout           time.Duration
	SkipSSLValidation bool
	ClientID          string
	ClientSecret      string
	TokenEndpoint     string
}

func New(config *Config) *OAuthClient {
	clientCredentialsConfig := &clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		AuthStyle:    oauth2.AuthStyleAutoDetect,
		TokenURL:     config.TokenEndpoint,
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, http.DefaultClient)
	authClient := clientCredentialsConfig.Client(ctx)

	return &OAuthClient{
		Client: authClient,
	}
}

func (c *OAuthClient) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}

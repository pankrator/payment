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
	Token(ctx context.Context) (*oauth2.Token, error)
}

type OAuthClient struct {
	*http.Client
	oauthConfig        *oauth2.Config
	tokenSourceCreator func(ctx context.Context) (oauth2.TokenSource, error)
}

type AuthFlow string

var (
	ClientCredentialsFlow AuthFlow = "client_credentials"
	PasswordFlow          AuthFlow = "password"
)

type Config struct {
	OauthURL          string
	Timeout           time.Duration
	SkipSSLValidation bool
	ClientID          string
	ClientSecret      string
	Username          string
	Password          string
	TokenEndpoint     string
	Flow              AuthFlow
	Token             *oauth2.Token
}

func New(config *Config) *OAuthClient {
	if config.TokenEndpoint == "" {
		info, err := GetInfo(config.OauthURL)
		if err != nil {
			// TODO: Fix this?!
			panic(err)
		}
		config.TokenEndpoint = info.TokenEndpoint
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL:  config.TokenEndpoint,
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}
	clientCredentialsConfig := &clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		AuthStyle:    oauth2.AuthStyleAutoDetect,
		TokenURL:     config.TokenEndpoint,
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, http.DefaultClient)
	authClient := clientCredentialsConfig.Client(ctx)

	if config.Flow == ClientCredentialsFlow {
		return &OAuthClient{
			Client: authClient,
			tokenSourceCreator: func(ctx context.Context) (oauth2.TokenSource, error) {
				return clientCredentialsConfig.TokenSource(ctx), nil
			},
		}
	} else if config.Flow == PasswordFlow {
		return &OAuthClient{
			Client: authClient,
			tokenSourceCreator: func(ctx context.Context) (oauth2.TokenSource, error) {
				if config.Token != nil && config.Token.RefreshToken != "" {
					return oauthConfig.TokenSource(ctx, config.Token), nil
				}
				token, err := oauthConfig.PasswordCredentialsToken(ctx, config.Username, config.Password)
				if err != nil {
					return nil, err
				}
				return oauthConfig.TokenSource(ctx, token), nil
			},
		}
	}

	return &OAuthClient{
		Client:      authClient,
		oauthConfig: oauthConfig,
	}
}

func (c *OAuthClient) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}

func (c *OAuthClient) Token(ctx context.Context) (*oauth2.Token, error) {
	tokenSource, err := c.tokenSourceCreator(ctx)
	if err != nil {
		return nil, err
	}
	return tokenSource.Token()
}

package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/pankrator/payment/auth"
	"github.com/pankrator/payment/web"
)

type UserData struct {
	Email  string
	Scopes []string
}

type Settings struct {
	OauthServerURL    string `mapstructure:"oauth_server_url"`
	AdminClientID     string `mapstructure:"admin_client_id"`
	AdminClientSecret string `mapstructure:"admin_client_secret"`
	ClientID          string `mapstructure:"client_id"`
	ClientSecret      string `mapstructure:"client_secret"`
}

func (s *Settings) Keys() []string {
	return []string{
		"oauth_server_url", "client_id", "client_secret",
		"admin_client_id", "admin_client_secret",
	}
}

func DefaultSettings() *Settings {
	return &Settings{}
}

type TokenAuthenticator struct {
	settings *Settings
	verifier *oidc.IDTokenVerifier
}

func NewTokenAuthenticator(ctx context.Context, settings *Settings) (*TokenAuthenticator, error) {
	info, err := auth.GetInfo(settings.OauthServerURL)
	if err != nil {
		return nil, fmt.Errorf("could not get oauth server info: %s", err)
	}

	keySet := oidc.NewRemoteKeySet(ctx, info.JWKsURI)
	return &TokenAuthenticator{
		verifier: oidc.NewVerifier(info.IssuerURL, keySet, &oidc.Config{
			ClientID: settings.ClientID,
		}),
		settings: settings,
	}, nil
}

func (ta *TokenAuthenticator) Authenticate(req *http.Request) (*UserData, error) {
	ctx := req.Context()
	authHeader := req.Header.Get("Authorization")
	if len(authHeader) < 7 {
		return nil, &web.HTTPError{
			StatusCode:  http.StatusUnauthorized,
			Description: "Bearer token not provided",
		}
	}
	tokenText := authHeader[len("Bearer "):]
	token, err := ta.verifier.Verify(ctx, tokenText)
	if err != nil {
		return nil, err
	}
	claims := &struct {
		Email  string   `json:"email"`
		Scopes []string `json:"scopes"`
	}{}
	if err := token.Claims(claims); err != nil {
		return nil, err
	}
	return &UserData{
		Email:  claims.Email,
		Scopes: claims.Scopes,
	}, nil
}

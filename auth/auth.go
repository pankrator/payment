package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/pankrator/payment/web"
)

var NoTokenProvidedErr = errors.New("no token provided")
var VerificationFailedErr = errors.New("could not verify token")

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
	info, err := GetInfo(settings.OauthServerURL)
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

func (ta *TokenAuthenticator) Authenticate(req *http.Request) (*web.UserData, error) {
	ctx := req.Context()
	authHeader := req.Header.Get("Authorization")
	if len(authHeader) < 7 {
		return nil, NoTokenProvidedErr
	}
	tokenText := authHeader[len("Bearer "):]

	token, err := ta.verifier.Verify(ctx, tokenText)
	if err != nil {
		log.Printf("could not verify token: %s", err)
		return nil, VerificationFailedErr
	}
	claims := &struct {
		Email    string   `json:"email"`
		Scopes   []string `json:"scope"`
		UserName string   `json:"user_name"`
	}{}
	if err := token.Claims(claims); err != nil {
		return nil, err
	}
	return &web.UserData{
		Email:  claims.Email,
		Scopes: claims.Scopes,
		Name:   claims.UserName,
	}, nil
}

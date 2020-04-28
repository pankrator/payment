package auth

import (
	"fmt"
	"net/http"

	"github.com/pankrator/payment/web"
)

type AuthInfo struct {
	IssuerURL     string `json:"issuer"`
	TokenEndpoint string `json:"token_endpoint"`
	JWKsURI       string `json:"jwks_uri"`
}

func GetInfo(oauthUrl string) (*AuthInfo, error) {
	resp, err := http.DefaultClient.Get(oauthUrl + "/.well-known/openid-configuration")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status returned from Oauth Server: %d", resp.StatusCode)
	}
	info := &AuthInfo{}
	if err := web.BodyToObject(resp.Body, info); err != nil {
		return nil, fmt.Errorf("could not read response body: %s", err)
	}

	return info, nil
}

package filter

import (
	"log"
	"net/http"

	"github.com/pankrator/payment/auth"
	"github.com/pankrator/payment/web"
)

type Auth struct {
	authenticator *auth.TokenAuthenticator
}

func NewAuthFilter(authenticator *auth.TokenAuthenticator) *Auth {
	return &Auth{
		authenticator: authenticator,
	}
}

func (m *Auth) Execute(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	user, err := m.authenticator.Authenticate(req)
	if err != nil {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusUnauthorized,
			Description: err.Error(),
		})
		return
	}
	log.Printf("Logged in user is: %s", user.Email)
	ctx := req.Context()
	ctx = web.ContextWithUser(ctx, user)
	req = req.WithContext(ctx)

	next.ServeHTTP(rw, req)
}

func (m *Auth) Matchers() []web.Endpoint {
	return []web.Endpoint{
		{
			Path:   "/payment",
			Method: http.MethodPost,
		},
		{
			Path:   "/payment",
			Method: http.MethodGet,
		},
	}
}

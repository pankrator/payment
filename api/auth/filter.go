package auth

import (
	"log"
	"net/http"

	"github.com/pankrator/payment/web"
)

type Filter struct {
	authenticator *TokenAuthenticator
}

func NewFilter(authenticator *TokenAuthenticator) *Filter {
	return &Filter{
		authenticator: authenticator,
	}
}

func (m *Filter) Execute(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	user, err := m.authenticator.Authenticate(req)
	if err != nil {
		web.WriteError(rw, err)
		return
	}
	log.Printf("Logged in user is: %s", user.Email)
	ctx := req.Context()
	ctx = ContextWithUser(ctx, user)
	req = req.WithContext(ctx)

	next.ServeHTTP(rw, req)
}

func (m *Filter) Matcher() web.Endpoint {
	return web.Endpoint{
		Path:   "/payment",
		Method: http.MethodPost,
	}
}

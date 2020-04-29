package web

import (
	"fmt"
	"net/http"
	"strings"
)

func ScopeWrapper(handler HandlerFunc, scopesFunc func() []string) HandlerFunc {
	var scopes []string
	if scopesFunc != nil {
		scopes = scopesFunc()
	}
	return func(rw http.ResponseWriter, req *Request) {
		if len(scopes) == 0 {
			handler(rw, req)
			return
		}
		user, found := UserFromContext(req.Request.Context())
		if !found {
			WriteError(rw, &HTTPError{
				StatusCode:  http.StatusUnauthorized,
				Description: fmt.Sprintf("User not found"),
			})
			return
		}

		for _, requiredScope := range scopes {
			scopeFound := false
			for _, scope := range user.Scopes {
				if scope == requiredScope {
					scopeFound = true
					break
				}
			}
			if !scopeFound {
				WriteError(rw, &HTTPError{
					StatusCode:  http.StatusForbidden,
					Description: fmt.Sprintf("User scopes [%s] does not contain required scope %s", strings.Join(user.Scopes, ","), requiredScope),
				})
				return
			}
		}

		handler(rw, req)
	}
}

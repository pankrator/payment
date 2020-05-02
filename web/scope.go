package web

import (
	"fmt"
	"net/http"
	"strings"
)

func ScopeWrapper(handler HandlerFunc, scopesFunc func() []string) HandlerFunc {
	var requiredScopes []string
	if scopesFunc != nil {
		requiredScopes = scopesFunc()
	}
	return func(rw http.ResponseWriter, req *Request) {
		if len(requiredScopes) == 0 {
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

		matched, scope := HasScopes(user.Scopes, requiredScopes)
		if !matched {
			WriteError(rw, &HTTPError{
				StatusCode:  http.StatusForbidden,
				Description: fmt.Sprintf("User scopes [%s] does not contain required scope %s", strings.Join(user.Scopes, ","), scope),
			})
			return
		}

		handler(rw, req)
	}
}

func HasScopes(scopes []string, required []string) (bool, string) {
	for _, requiredScope := range required {
		scopeMatched := false
		for _, scope := range scopes {
			if scope == requiredScope {
				scopeMatched = true
				break
			}
		}
		if !scopeMatched {
			return false, requiredScope
		}
	}
	return true, ""
}

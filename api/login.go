package api

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/pankrator/payment/auth"
	"github.com/pankrator/payment/web"
	"golang.org/x/oauth2"
)

type LoginController struct {
	config *auth.Settings
}

func NewLoginController(authConfig *auth.Settings) web.Controller {
	return &LoginController{
		config: authConfig,
	}
}

func (c *LoginController) loginPage(rw http.ResponseWriter, req *web.Request) {
	fp := path.Join("templates", "login.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusInternalServerError,
			Description: "could not load view",
		})
		return
	}

	if err := tmpl.Execute(rw, nil); err != nil {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusInternalServerError,
			Description: "could not load view",
		})
		return
	}
}

func (c *LoginController) login(rw http.ResponseWriter, req *web.Request) {
	ctx := req.Request.Context()
	req.Request.ParseForm()
	username := req.Request.FormValue("username")
	password := req.Request.FormValue("password")

	client := auth.New(&auth.Config{
		OauthURL:          c.config.OauthServerURL,
		ClientID:          c.config.ClientID,
		ClientSecret:      c.config.ClientSecret,
		Flow:              auth.PasswordFlow,
		Timeout:           time.Second * 10,
		SkipSSLValidation: false,
		Username:          username,
		Password:          password,
	})

	token, err := client.Token(ctx)
	if err != nil {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusForbidden,
			Description: fmt.Sprintf("Could not get token: %s", err),
		})
		return
	}

	http.SetCookie(rw, &http.Cookie{
		Expires:  time.Now().Add(time.Hour),
		HttpOnly: true,
		Name:     "refresh_token",
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Value:    token.RefreshToken,
	})
	web.WriteJSON(rw, http.StatusOK, token.AccessToken)
}

func (c *LoginController) refresh(rw http.ResponseWriter, req *web.Request) {
	log.Printf("Refresh token endpoint called")
	ctx := req.Request.Context()
	cookie, err := req.Request.Cookie("refresh_token")
	if err == http.ErrNoCookie {
		web.WriteJSON(rw, http.StatusNotFound, map[string]interface{}{})
		return
	}

	refreshToken := cookie.Value
	client := auth.New(&auth.Config{
		OauthURL:          c.config.OauthServerURL,
		ClientID:          c.config.ClientID,
		ClientSecret:      c.config.ClientSecret,
		Flow:              auth.PasswordFlow,
		Timeout:           time.Second * 10,
		SkipSSLValidation: false,
		Token: &oauth2.Token{
			RefreshToken: refreshToken,
		},
	})
	token, err := client.Token(ctx)
	if err != nil {
		web.WriteError(rw, fmt.Errorf("could not refresh token: %s", err))
		return
	}
	http.SetCookie(rw, &http.Cookie{
		Expires:  time.Now().Add(time.Minute * 60),
		HttpOnly: true,
		Name:     "refresh_token",
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Value:    token.RefreshToken,
	})
	web.WriteJSON(rw, http.StatusOK, token.AccessToken)
}

func (c *LoginController) logout(rw http.ResponseWriter, req *web.Request) {
	log.Printf("Logout endpoint called")

	http.SetCookie(rw, &http.Cookie{
		Expires:  time.Now().Add(time.Minute * 60),
		HttpOnly: true,
		Name:     "refresh_token",
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Value:    "",
	})
	http.Redirect(rw, req.Request, "/", http.StatusMovedPermanently)
}

func (c *LoginController) Routes() []web.Route {
	return []web.Route{
		{
			Endpoint: web.Endpoint{
				Method: http.MethodGet,
				Path:   "/login",
			},
			Handler: c.loginPage,
		},
		{
			Endpoint: web.Endpoint{
				Method: http.MethodGet,
				Path:   "/refresh",
			},
			Handler: c.refresh,
		},
		{
			Endpoint: web.Endpoint{
				Method: http.MethodPost,
				Path:   "/login",
			},
			Handler: c.login,
		},
		{
			Endpoint: web.Endpoint{
				Method: http.MethodGet,
				Path:   "/logout",
			},
			Handler: c.logout,
		},
	}
}

package test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"path"
	"runtime"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gavv/httpexpect"
	"github.com/pankrator/payment/app"
	"github.com/pankrator/payment/auth"
	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/users"
)

type TestApp struct {
	Server         *httptest.Server
	Repository     storage.Storage
	Expect         *httpexpect.Expect
	ExpectWithAuth *httpexpect.Expect
}

func NewTestApp() *TestApp {
	_, b, _, _ := runtime.Caller(0)
	basePath := path.Dir(b)

	paymentApp := app.New(basePath)
	testServer := httptest.NewUnstartedServer(paymentApp.Server.Router)
	testServer.Start()
	if err := paymentApp.Repository.Open(func(driver, url string) (*sql.DB, error) {
		return sql.Open(driver, url)
	}); err != nil {
		panic(err)
	}
	paymentApp.Settings.Users.FileLocation = basePath

	userInitiator := app.NewUserInitiator(paymentApp.UaaClient, paymentApp.Repository, paymentApp.MerchantService)
	usersByType := userInitiator.LoadUsers(paymentApp.Settings.Users)

	adminGroups := []string{"merchant.read", "merchant.write", "merchant.delete", "transaction.write", "transaction.read"}
	merchantGroups := []string{"transaction.read"}

	userInitiator.InitUsers(
		context.TODO(),
		usersByType[users.Admin],
		adminGroups,
	)
	userInitiator.InitUsers(
		context.TODO(),
		usersByType[users.Merchant],
		merchantGroups,
	)

	authSettings := paymentApp.Settings.Auth
	client := auth.New(&auth.Config{
		ClientID:     authSettings.ClientID,
		ClientSecret: authSettings.ClientSecret,
		OauthURL:     authSettings.OauthServerURL,
		Username:     "admin",
		Password:     "1234",
		Flow:         auth.PasswordFlow,
	})

	oauthToken, err := client.Token(context.TODO())
	Expect(err).ShouldNot(HaveOccurred())

	expect := httpexpect.New(ginkgo.GinkgoT(), testServer.URL)
	expectWithAuth := expect.Builder(func(req *httpexpect.Request) {
		req.WithHeader("Authorization", fmt.Sprintf("Bearer %s", oauthToken.AccessToken))
	})
	result := expectWithAuth.GET("/transactions").Expect()
	secureCookie := result.Cookie("_gorilla_csrf").Value().Raw()
	csrfToken := result.Header("X-CSRF-Token").Raw()
	expectWithAuth = expectWithAuth.Builder(func(req *httpexpect.Request) {
		// req.WithHeader("Authorization", fmt.Sprintf("Bearer %s", oauthToken.AccessToken))
		req.WithHeader("X-CSRF-Token", csrfToken)
		req.WithCookie("_gorilla_csrf", secureCookie)
	})

	return &TestApp{
		Server:         testServer,
		Repository:     paymentApp.Repository,
		Expect:         expect,
		ExpectWithAuth: expectWithAuth,
	}
}

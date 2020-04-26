package test

import (
	"database/sql"
	"net/http/httptest"
	"path"
	"runtime"

	"github.com/gavv/httpexpect"
	"github.com/onsi/ginkgo"
	"github.com/pankrator/payment/app"
	"github.com/pankrator/payment/storage"
)

type TestApp struct {
	Server     *httptest.Server
	Repository storage.Storage
	Expect     *httpexpect.Expect
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

	expect := httpexpect.New(ginkgo.GinkgoT(), testServer.URL)

	return &TestApp{
		Server:     testServer,
		Repository: paymentApp.Repository,
		Expect:     expect,
	}
}

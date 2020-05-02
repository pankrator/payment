package web_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/gavv/httpexpect"

	. "github.com/onsi/ginkgo"
	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/web"
)

var _ = Describe("Server", func() {

	web.RegisterParser("application/xml", &web.XMLParser{})
	web.RegisterParser("application/json", &web.JSONParser{})

	settings := web.DefaultSettings()
	settings.UseCSRFProtection = false
	server := web.NewServer(settings, &web.Api{
		Controllers: []web.Controller{&testCtrl{}},
	})
	testServer := httptest.NewServer(server.Router)
	psExpect := httpexpect.New(GinkgoT(), testServer.URL)

	Context("registered path", func() {
		Context("GET", func() {
			It("should return data", func() {
				psExpect.GET("/test_get").Expect().
					Status(http.StatusOK).
					JSON().Object().
					Equal(map[string]interface{}{"result": "OK"})
			})
		})

		Context("POST", func() {
			Context("with valid json", func() {
				It("should return the input", func() {
					input := &model.Transaction{
						MerchantID: "0",
						Type:       "authorize",
						Amount:     10,
					}
					psExpect.POST("/test_post").WithJSON(input).Expect().
						Status(http.StatusCreated).
						JSON().Object().
						Equal(input)
				})

				Context("for invalid model", func() {
					It("should return error", func() {
						input := &model.Transaction{
							Amount: 10,
						}
						psExpect.POST("/test_post").WithJSON(input).Expect().
							Status(http.StatusBadRequest).
							JSON().Object().Value("description").String().
							Contains("Validation of model failed: merchant id is required for transaction")
					})
				})
			})

			Context("with valid xml", func() {
				It("should return the input", func() {
					input := &model.Transaction{
						MerchantID: "0",
						Type:       "authorize",
						Amount:     10,
					}
					psExpect.POST("/test_post").
						WithHeader("Content-Type", "application/xml").
						WithBytes([]byte(`<Transaction><Type>authorize</Type><Amount>10</Amount><MerchantID>0</MerchantID></Transaction>`)).
						Expect().
						Status(http.StatusCreated).
						JSON().Object().
						Equal(input)
				})
			})

			Context("with invalid json", func() {
				It("should return an error", func() {
					psExpect.POST("/test_post").WithBytes([]byte(`{"wrong:json"}`)).Expect().
						Status(http.StatusBadRequest).
						JSON().Object().Value("description").String().Contains("No parser found for type")
				})
			})
		})
	})

	Context("unknown path", func() {
		It("should return not found", func() {
			psExpect.GET("/unknown").Expect().Status(http.StatusNotFound)
		})
	})

	When("controller panics", func() {
		It("should recover", func() {
			psExpect.GET("/panic").Expect().Status(http.StatusInternalServerError).
				JSON().Object().Value("description").
				String().Contains("Unexpected error occured")
		})
	})
})

type testCtrl struct{}

func (t *testCtrl) Routes() []web.Route {
	return []web.Route{
		{
			Endpoint: web.Endpoint{
				Method: http.MethodGet,
				Path:   "/test_get",
			},
			Handler: func(rw http.ResponseWriter, req *web.Request) {
				web.WriteJSON(rw, http.StatusOK, map[string]interface{}{
					"result": "OK",
				})
			},
		},
		{
			Endpoint: web.Endpoint{
				Method: http.MethodPost,
				Path:   "/test_post",
			},
			Handler: func(rw http.ResponseWriter, req *web.Request) {
				web.WriteJSON(rw, http.StatusCreated, req.Model)
			},
			ModelBlueprint: func() model.Object {
				return &model.Transaction{}
			},
		},
		{
			Endpoint: web.Endpoint{
				Method: http.MethodGet,
				Path:   "/panic",
			},
			Handler: func(rw http.ResponseWriter, req *web.Request) {
				panic("unexpected")
			},
		},
	}
}

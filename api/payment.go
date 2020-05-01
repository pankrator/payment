package api

import (
	"html/template"
	"log"
	"net/http"
	"path"

	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/query"

	"github.com/pankrator/payment/web"
)

type PaymentService interface {
	Create(*model.Transaction) (model.Object, error)
	List(query []query.Query) ([]model.Object, error)
}

type PaymentController struct {
	paymentService PaymentService
}

func NewPaymentController(paymentService PaymentService) web.Controller {
	return &PaymentController{
		paymentService: paymentService,
	}
}

func (c *PaymentController) payment(rw http.ResponseWriter, req *web.Request) {
	log.Printf("Received payment transaction %+v", req.Model)

	transaction := req.Model.(*model.Transaction)

	result, err := c.paymentService.Create(transaction)
	if err != nil {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusBadRequest,
			Description: err.Error(),
		})
		return
	}

	web.WriteJSON(rw, http.StatusCreated, result)
}

func (c *PaymentController) list(rw http.ResponseWriter, req *web.Request) {
	q := query.QueryFromContext(req.Request.Context())
	result, err := c.paymentService.List(q)
	if err != nil {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusBadRequest,
			Description: err.Error(),
		})
		return
	}
	web.WriteJSON(rw, http.StatusCreated, result)
}

func (c *PaymentController) view(rw http.ResponseWriter, req *web.Request) {
	fp := path.Join("templates", "payments.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusInternalServerError,
			Description: "could not load view",
		})
		return
	}
	q := query.QueryFromContext(req.Request.Context())
	result, err := c.paymentService.List(q)
	if err != nil {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusBadRequest,
			Description: err.Error(),
		})
		return
	}

	if err := tmpl.Execute(rw, result); err != nil {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusInternalServerError,
			Description: "could not load view",
		})
		return
	}
}

func (c *PaymentController) Routes() []web.Route {
	return []web.Route{
		{
			ModelBlueprint: func() model.Object {
				return &model.Transaction{}
			},
			Endpoint: web.Endpoint{
				Method: http.MethodPost,
				Path:   "/payment",
			},
			Scopes: func() []string {
				return []string{"transaction.write"}
			},
			Handler: c.payment,
		},
		{
			Endpoint: web.Endpoint{
				Method: http.MethodGet,
				Path:   "/payment",
			},
			Scopes: func() []string {
				return []string{"transaction.read"}
			},
			Handler: c.list,
		},
		{
			Endpoint: web.Endpoint{
				Method: http.MethodGet,
				Path:   "/payment/view",
			},
			Handler: c.view,
			Scopes: func() []string {
				return []string{"transaction.read"}
			},
		},
	}
}

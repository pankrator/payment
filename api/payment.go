package api

import (
	"log"
	"net/http"

	"github.com/pankrator/payment/model"

	"github.com/pankrator/payment/web"
)

type PaymentService interface {
	Create(*model.Transaction) (model.Object, error)
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
	}
}

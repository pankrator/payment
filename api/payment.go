package api

import (
	"log"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/pankrator/payment/model"

	"github.com/pankrator/payment/web"
)

type PaymentService interface {
	Create(model.Object) (model.Object, error)
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

	UUID, err := uuid.NewV4()
	if err != nil {
		log.Printf("Could not generate UUID: %s", err)
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusInternalServerError,
			Description: "Internal error",
		})
		return
	}

	transaction := req.Model.(*model.Transaction)
	transaction.UUID = UUID.String()

	result, err := c.paymentService.Create(req.Model)
	if err != nil {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusBadRequest,
			Description: err.Error(),
		})
		return
	}

	rw.WriteHeader(http.StatusCreated)
	web.WriteJSON(rw, result)
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
			Handler: c.payment,
		},
	}
}

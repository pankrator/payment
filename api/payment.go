package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/pankrator/payment/api/auth"
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
	user, found := auth.UserFromContext(req.Request.Context())
	if !found {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusUnauthorized,
			Description: fmt.Sprintf("User not found"),
		})
		return
	}
	// TODO: Extract scope checks
	scopeFound := false
	for _, scope := range user.Scopes {
		if scope == "transaction.write" {
			scopeFound = true
		}
	}
	if !scopeFound {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusForbidden,
			Description: fmt.Sprintf("User scopes [%s] does not contain required scope transaction.write", strings.Join(user.Scopes, ",")),
		})
	}

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
			Handler: c.payment,
		},
	}
}

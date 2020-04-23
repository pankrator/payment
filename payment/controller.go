package payment

import (
	"log"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/pankrator/payment/model"

	"github.com/pankrator/payment/storage"

	"github.com/pankrator/payment/web"
)

type Controller struct {
	Repository *storage.Storage
}

func (c *Controller) payment(rw http.ResponseWriter, req *web.Request) {
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

	result, err := c.Repository.Create(transaction)
	if err != nil {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusBadRequest,
			Description: err.Error(),
		})
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	web.WriteJSON(rw, result)
}

func (c *Controller) Routes() []web.Route {
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

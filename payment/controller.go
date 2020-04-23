package payment

import (
	"encoding/json"
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
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	transaction := req.Model.(*model.Transaction)
	transaction.UUID = UUID.String()

	result, err := c.Repository.Create(transaction)
	// TODO: Use generic methods to write errors
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	bytes, err := json.Marshal(result)
	// TODO: Use generic methods to write errors
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	rw.Write(bytes)
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

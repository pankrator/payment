package payment

import (
	"fmt"
	"net/http"

	"github.com/pankrator/payment/web"
)

type Controller struct{}

func (c *Controller) payment(rw http.ResponseWriter, req *web.Request) {
	fmt.Printf("%+v\n", req.Model)
	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte("{}"))
}

func (c *Controller) Routes() []web.Route {
	return []web.Route{
		{
			ModelBlueprint: func() interface{} {
				return &Transaction{}
			},
			Endpoint: web.Endpoint{
				Method: http.MethodPost,
				Path:   "/payment",
			},
			Handler: c.payment,
		},
	}
}

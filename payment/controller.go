package payment

import (
	"net/http"

	"github.com/pankrator/payment/web"
)

type Controller struct{}

func (c *Controller) payment(rw http.ResponseWriter, req *http.Request) {
}

func (c *Controller) Routes() []web.Route {
	return []web.Route{
		{
			Endpoint: web.Endpoint{
				Method: http.MethodPost,
				Path:   "/payment",
			},
			Handler: c.payment,
		},
	}
}

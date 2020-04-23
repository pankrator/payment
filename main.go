package main

import (
	"context"

	"github.com/pankrator/payment/payment"
	"github.com/pankrator/payment/web"
)

func main() {
	api := &web.Api{
		Controllers: []web.Controller{
			&payment.Controller{},
		},
	}
	server := web.NewServer(web.DefaultSettings(), api)

	server.Run(context.Background())
}

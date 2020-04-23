package main

import (
	"context"

	"github.com/pankrator/payment/app"
)

func main() {
	application := app.New()
	application.Start(context.Background())
}

package main

import (
	"context"

	"github.com/pankrator/payment/app"
)

func main() {
	application := app.New(".")
	application.InitUsers(context.Background())
	application.Start(context.Background())
}

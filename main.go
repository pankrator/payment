package main

import (
	"context"
	"log"
	"sync"

	"github.com/pankrator/payment/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	application := app.New(".")
	application.Start(ctx, wg, cancel)
	wg.Wait()
	log.Printf("Everything closed. Closing the process")
}

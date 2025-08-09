package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/MatTwix/Food-Delivery-Agregator/notification-service/messaging"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	messaging.StartConsumers(ctx)

	log.Println("Notificaion service started. Waiting for events...")
	<-ctx.Done()
	log.Printf("Notificaions service shutting down.")
}

package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/MatTwix/Food-Delivery-Agregator/payments-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/payments-service/messaging"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config.InitConfig()
	messaging.InitTopicsNames()

	messaging.InitTopics()

	producer, err := messaging.NewProducer()
	if err != nil {
		log.Fatalf("Error creating producer: %v", err)
	}
	defer producer.Close()

	messaging.StartConsumers(ctx, producer)

	log.Println("Payments service started. Waiting for events...")
	<-ctx.Done()
	log.Println("Payments service shutting down.")
}

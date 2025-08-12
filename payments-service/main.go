package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/MatTwix/Food-Delivery-Agregator/payments-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/payments-service/messaging"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config.InitConfig()
	config.InitLogger()

	messaging.InitTopicsNames()

	messaging.InitTopics()

	producer, err := messaging.NewProducer()
	if err != nil {
		slog.Error("failed to create producer", "error", err)
		os.Exit(1)
	}
	defer producer.Close()

	messaging.StartConsumers(ctx, producer)

	slog.Info("payments service started. Waiting for events...")
	<-ctx.Done()
	slog.Info("payments service shutting down")
}

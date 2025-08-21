package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/MatTwix/Food-Delivery-Agregator/notification-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/notification-service/messaging"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config.InitConfig()
	config.InitLogger()

	messaging.InitTopicsNames()

	messaging.InitTopics()

	kafkaProducer, err := messaging.NewProducer()
	if err != nil {
		slog.Error("failed to create Kafka producer", "error", err)
		os.Exit(1)
	}

	messaging.StartConsumers(ctx, kafkaProducer)

	slog.Info("notificaion service started. Waiting for events...")
	<-ctx.Done()
	slog.Info("notificaions service shutting down")
}

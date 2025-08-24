package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/MatTwix/Food-Delivery-Agregator/scheduler-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/scheduler-service/messaging"
)

func main() {
	config.InitConfig()
	config.InitLogger()

	messaging.InitTopicsNames()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	messaging.InitTopics()

	kafkaProducer, err := messaging.NewProducer()
	if err != nil {
		slog.Error("failed to create Kafka producer", "error", err)
		os.Exit(1)
	}

	// orderGRPCClient := clients.NewOrdersSerciceClient()

	<-ctx.Done()

	slog.Info("shutting down servers")

	kafkaProducer.Close()
	slog.Info("Kafka producer closed")

	slog.Info("service gracefully stopped")
}

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/MatTwix/Food-Delivery-Agregator/scheduler-service/clients"
	"github.com/MatTwix/Food-Delivery-Agregator/scheduler-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/scheduler-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/scheduler-service/scheduler"
	"github.com/robfig/cron/v3"
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

	orderGRPCClient := clients.NewOrdersSerciceClient()
	c := cron.New()

	requestCourierJob := scheduler.NewRequestCourierJob(orderGRPCClient, kafkaProducer)
	scheduler.RegisterJobs(c, requestCourierJob)

	go c.Run()

	<-ctx.Done()

	slog.Info("shutting down servers")

	c.Stop()
	slog.Info("cron stopped")

	kafkaProducer.Close()
	slog.Info("Kafka producer closed")

	slog.Info("service gracefully stopped")
}

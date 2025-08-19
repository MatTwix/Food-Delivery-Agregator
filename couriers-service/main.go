package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/api"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/database"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/store"
)

func main() {
	config.InitConfig()
	config.InitValidator()
	config.InitLogger()

	messaging.InitTopicsNames()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database.NewConnection()
	db := database.DB

	messaging.InitTopics()

	courierStore := store.NewCourierStore(db)
	deliveryStore := store.NewDeliveryStore(db)

	kafkaProducer, err := messaging.NewProducer()
	if err != nil {
		slog.Error("failed to creat Kafka producer", "error", err)
		os.Exit(1)
	}

	router := api.SetupRoutes(deliveryStore, courierStore, kafkaProducer)
	httpServer := &http.Server{
		Addr:    ":" + config.Cfg.HTTP.Port,
		Handler: router,
	}

	messaging.StartConsumers(ctx, courierStore, deliveryStore, kafkaProducer)

	go func() {
		slog.Info("starting couriers service", "port", config.Cfg.HTTP.Port)

		if err := httpServer.ListenAndServe(); err != nil {
			slog.Error("failed to start service", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	slog.Info("shutting down servers...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shut down servers", "error", err)
	}
	slog.Info("HTTP server stopped")

	kafkaProducer.Close()
	slog.Info("Kafka producer closed")

	database.DB.Close()
	slog.Info("database connection closed")

	slog.Info("service gracefully stopped")
}

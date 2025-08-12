package main

import (
	"context"
	"log"
	"net/http"
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
		log.Fatalf("Error creating Kafka producer: %v", err)
	}

	router := api.SetupRoutes(courierStore, kafkaProducer)
	httpServer := &http.Server{
		Addr:    ":" + config.Cfg.HTTP.Port,
		Handler: router,
	}

	messaging.StartConsumers(ctx, courierStore, deliveryStore, kafkaProducer)

	go func() {
		log.Printf("Starting couriers service on port %s", config.Cfg.HTTP.Port)

		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start service: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down servers...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down servers: %v", err)
	}
	log.Println("HTTP server stopped.")

	kafkaProducer.Close()
	log.Println("Kafka producer closed.")

	database.DB.Close()
	log.Println("Database connection closed.")

	log.Println("Service gracefully stopped.")
}

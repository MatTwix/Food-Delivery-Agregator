package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/api"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/clients"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/database"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/store"
)

func main() {
	cfg := config.LoadConfig()
	config.InitValidator()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database.NewConnection()
	db := database.DB

	messaging.InitTopics()

	restaurantStore := store.NewRestaurantStore(db)
	orderStore := store.NewOrderStore(db)

	kafkaProducer, err := messaging.NewProducer()
	if err != nil {
		log.Fatalf("Error creating producer: %v", err)
	}

	grpcClient := clients.NewResraurantServiceClient()

	router := api.SetupRoutes(restaurantStore, orderStore, grpcClient, kafkaProducer)
	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	messaging.StartConsumers(ctx, restaurantStore, orderStore, kafkaProducer)

	go func() {
		log.Printf("Starting orders service on port %s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
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

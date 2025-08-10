package main

import (
	"context"
	"log"
	"net/http"
	"os"
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

	messaging.InitTopics()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database.NewConnection()
	db := database.DB

	restaurantStore := store.NewRestaurantStore(db)
	orderStore := store.NewOrderStore(db)

	producer, err := messaging.NewProducer()
	if err != nil {
		log.Fatalf("Error creating producer: %v", err)
	}
	defer producer.Close()
	grpcClient := clients.NewResraurantServiceClient()

	messaging.StartConsumers(ctx, restaurantStore, orderStore, producer)

	router := api.SetupRoutes(restaurantStore, orderStore, grpcClient, producer)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Starting orders service on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Received shutdown signal, gracefully shutting down...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Service shut down gracefully")
}

package main

import (
	"context"
	"log"
	"net/http"

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

	log.Printf("Starting orders service on port %s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

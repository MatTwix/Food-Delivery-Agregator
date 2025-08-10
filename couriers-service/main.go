package main

import (
	"context"
	"log"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/api"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/database"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/store"
)

func main() {
	cfg := config.LoadConfig()

	config.InitValidator()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database.NewConnection()
	db := database.DB
	defer db.Close()

	courierStore := store.NewCourierStore(db)
	deliveryStore := store.NewDeliveryStore(db)

	producer, err := messaging.NewProducer()
	if err != nil {
		log.Fatalf("Error creating producer: %v", err)
	}
	defer producer.Close()

	messaging.StartConsumers(ctx, courierStore, deliveryStore, producer)

	r := api.SetupRoutes(courierStore, producer)

	log.Printf("Starting couriers service on port %s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}
}

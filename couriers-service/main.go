package main

import (
	"log"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/api"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/database"
)

func main() {
	cfg := config.LoadConfig()

	config.InitValidator()

	database.NewConnection()
	db := database.DB
	defer db.Close()

	r := api.SetupRoutes(db)

	log.Printf("Starting couriers service on port %s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}
}

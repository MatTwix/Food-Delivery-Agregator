package main

import (
	"log"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/database"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/router"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.LoadConfig()
	dbPool := database.NewConnection()
	defer dbPool.Close()

	r := chi.NewRouter()
	router.SetupRoutes(r)

	log.Printf("Starting restaurants service on port %s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}
}

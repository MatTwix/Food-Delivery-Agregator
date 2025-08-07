package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/clients"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/handlers"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRoutes(db *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	kafkaProducer, err := messaging.NewProducer()
	if err != nil {
		log.Fatalf("Error starting Kafka producer: %v", err)
	}

	restaurantStore := store.NewRestaurantStore(db)
	orderStore := store.NewOrderStore(db)
	grpcClient := clients.NewResraurantServiceClient()

	orderHandler := handlers.NewOrderHandler(orderStore, restaurantStore, grpcClient, kafkaProducer)

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", orderHandler.CreateOrder)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Orders service is up and running!")
	})

	return r
}

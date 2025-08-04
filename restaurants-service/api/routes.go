package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/handlers"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/store"
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
		log.Fatalf("Error creating Kafka producer: %v", err)
	}

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Restaurants service is up and running!")
	})

	restaurantStore := store.NewRestaurantsStore(db)
	restaurantHandler := handlers.NewRestaurantHandler(restaurantStore, kafkaProducer)

	r.Route("/restaurants", func(r chi.Router) {
		r.Get("/", restaurantHandler.GetRestaurants)
		r.Get("/{id}", restaurantHandler.GetRestaurantByID)
		r.Post("/", restaurantHandler.CreateRestaurant)
		r.Put("/{id}", restaurantHandler.UpdateRestaurant)
		r.Delete("/{id}", restaurantHandler.DeleteRestaurant)
	})

	return r
}

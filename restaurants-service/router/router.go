package router

import (
	"fmt"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/handlers"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRoutes(db *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Restaurants service is up and running!")
	})

	restaurantStore := store.NewRestaurantsStore(db)
	restaurantHandler := handlers.NewRestaurantHandler(restaurantStore)

	r.Route("/restaurants", func(r chi.Router) {
		r.Post("/", restaurantHandler.CreateRestaurant)
	})

	return r
}

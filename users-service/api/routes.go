package api

import (
	"fmt"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/users-service/handlers"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(userStore *store.UserStore) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Users service is up and running!")
	})

	userHandler := handlers.NewUserHandler(userStore)

	r.Route("/users", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
	})

	return r
}

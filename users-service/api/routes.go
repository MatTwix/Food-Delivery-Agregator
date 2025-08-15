package api

import (
	"fmt"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/users-service/handlers"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(userStore *store.UserStore, tokenStore *store.TokenStore) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Users service is up and running!")
	})

	userHandler := handlers.NewUserHandler(userStore, tokenStore)

	r.Route("/", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)
		r.Post("/refresh", userHandler.Refresh)
	})

	r.Route("/users", func(r chi.Router) {
		r.Get("/", userHandler.GetAllUses)
	})

	return r
}

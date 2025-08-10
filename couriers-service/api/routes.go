package api

import (
	"fmt"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/handlers"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/store"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(couriersStore *store.CourierStore, producer *messaging.Producer) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Couriers service is up and running!")
	})

	couriersHandler := handlers.NewCourierHandler(couriersStore, producer)

	r.Route("/couriers", func(r chi.Router) {
		r.Get("/", couriersHandler.GetCouriers)
		r.Get("/available", couriersHandler.GetAvailableCourier)
		r.Post("/", couriersHandler.CreateCourier)
		r.Post("/{orderID}/delivered", couriersHandler.DeliverOrder)
		r.Put("/{id}", couriersHandler.UpdateCourier)
		r.Delete("/{id}", couriersHandler.DeleteCourier)
	})

	return r
}

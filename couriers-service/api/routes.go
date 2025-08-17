package api

import (
	"fmt"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/common/auth"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/handlers"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/middleware"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/store"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(couriersStore *store.CourierStore, producer *messaging.Producer) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Couriers service is up and running!")
	})

	couriersHandler := handlers.NewCourierHandler(couriersStore, producer)

	r.Route("/couriers", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authorize("admin"))

			r.Get("/", couriersHandler.GetCouriers)
			r.Get("/available", couriersHandler.GetAvailableCourier)
			r.Post("/", couriersHandler.CreateCourier)
			r.Put("/{id}", couriersHandler.UpdateCourier)
			r.Delete("/{id}", couriersHandler.DeleteCourier)
		})

	})

	r.Route("/orders", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authorize(auth.RoleAdmin, auth.RoleCourier))

			r.Post("/{orderID}/picked_up", couriersHandler.PickUpOrder)
			r.Post("/{orderID}/delivered", couriersHandler.DeliverOrder)
		})
	})

	return r
}

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

func SetupRoutes(deliveryStore *store.DeliveryStore, courierStore *store.CourierStore, producer *messaging.Producer) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Couriers service is up and running!")
	})

	couriersHandler := handlers.NewCourierHandler(courierStore, producer)

	r.Route("/couriers", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authorize(auth.RoleAdmin))

			r.Get("/", couriersHandler.GetCouriers)
			r.Get("/available", couriersHandler.GetAvailableCourier)
			r.Put("/{id}", couriersHandler.UpdateCourier)
			// r.Delete("/{id}", couriersHandler.DeleteCourier)
		})

	})

	r.Route("/orders", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthorizeOwnerOrRoles(deliveryStore.GetPerformerID, auth.RoleAdmin, auth.RoleCourier))

			r.Post("/{id}/picked_up", couriersHandler.PickUpOrder)
			r.Post("/{id}/delivered", couriersHandler.DeliverOrder)
		})
	})

	return r
}

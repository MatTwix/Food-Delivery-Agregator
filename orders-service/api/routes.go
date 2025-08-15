package api

import (
	"fmt"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/common/auth"
	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/handlers"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/middleware"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/store"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(restaurantStore *store.RestaurantStore, orderStore *store.OrderStore, grpcClient pb.RestaurantServiceClient, kafkaProducer *messaging.Producer) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	orderHandler := handlers.NewOrderHandler(orderStore, restaurantStore, grpcClient, kafkaProducer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Orders service is up and running!")
	})

	r.Route("/orders", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.CheckRole(auth.RoleManager))
			r.Get("/", orderHandler.GetAllOrders)
		})
		r.Get("/{id}", orderHandler.GetOrderByID)
		r.Post("/", orderHandler.CreateOrder)
	})

	return r
}

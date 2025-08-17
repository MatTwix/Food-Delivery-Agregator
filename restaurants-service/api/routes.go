package api

import (
	"fmt"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/common/auth"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/handlers"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/middleware"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/store"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(restaurantStore *store.RestaurantStore, menuItemStore *store.MenuItemStore, kafkaProducer *messaging.Producer) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Restaurants service is up and running!")
	})

	restaurantHandler := handlers.NewRestaurantHandler(restaurantStore, kafkaProducer)

	r.Route("/restaurants", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authorize(auth.RoleAdmin, auth.RoleManager))
			r.Post("/", restaurantHandler.CreateRestaurant)
		})

		r.Get("/", restaurantHandler.GetRestaurants)
		r.Get("/{id}", restaurantHandler.GetRestaurantByID)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthorizeOwnerOrRoles(restaurantStore.GetOwnerID, auth.RoleAdmin, auth.RoleManager))

			r.Put("/{id}", restaurantHandler.UpdateRestaurant)
			r.Delete("/{id}", restaurantHandler.DeleteRestaurant)
		})
	})

	menuItemHandler := handlers.NewMenuItemHandler(menuItemStore)

	r.Route("/menu_items", func(r chi.Router) {
		r.Get("/", menuItemHandler.GetMenuItems)
		r.Get("/restaurant/{id}", menuItemHandler.GetMenuItemsByRestaurantID)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthorizeOwnerOrRoles(restaurantStore.GetOwnerID, auth.RoleAdmin, auth.RoleManager))

			r.Post("/restaurant/{id}", menuItemHandler.CreateMenuItem)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthorizeOwnerOrRoles(menuItemStore.GetRestaurantOwnerID, auth.RoleAdmin, auth.RoleManager))

			r.Put("/{id}", menuItemHandler.UpdateMenuItem)
			r.Delete("/{id}", menuItemHandler.DeleteMenuItem)
		})
	})

	return r
}

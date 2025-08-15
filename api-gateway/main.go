package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/api-gateway/config"
	"github.com/MatTwix/Food-Delivery-Agregator/api-gateway/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config.InitConfig()
	config.InitLogger()

	if config.Cfg.URLs.RestaurantsService == "" {
		slog.Error("RESTAURANTS_SERVICE_URL is not set")
		os.Exit(1)
	}
	if config.Cfg.URLs.OrdersService == "" {
		slog.Error("ORDERS_SERVICE_URL is not set")
		os.Exit(1)
	}
	if config.Cfg.URLs.CouriersService == "" {
		slog.Error("COURIERS_SERVICE_URL is not set")
		os.Exit(1)
	}
	if config.Cfg.URLs.UsersService == "" {
		slog.Error("USERS_SERVICE_URL is not set")
		os.Exit(1)
	}

	restaurantsProxy := createReverseProxy(config.Cfg.URLs.RestaurantsService)
	ordersProxy := createReverseProxy(config.Cfg.URLs.OrdersService)
	couriersProxy := createReverseProxy(config.Cfg.URLs.CouriersService)
	usersProxy := createReverseProxy(config.Cfg.URLs.UsersService)

	restaurantsProxyHandler := http.StripPrefix("/api/restaurants", restaurantsProxy)
	ordersProxyHandler := http.StripPrefix("/api/orders", ordersProxy)
	couriersProxyHandler := http.StripPrefix("/api/couriers", couriersProxy)
	usersProxyHandler := http.StripPrefix("/api/users", usersProxy)

	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		slog.Info("setting up public routes")
		r.Post("/api/users/register", usersProxyHandler.ServeHTTP)
		r.Post("/api/users/login", usersProxyHandler.ServeHTTP)
		r.Post("/api/users/refresh", usersProxyHandler.ServeHTTP)

		r.Get("/api/restaurants/restaurants", restaurantsProxyHandler.ServeHTTP)
		r.Get("/api/restaurants/restaurants/{id}", restaurantsProxyHandler.ServeHTTP)
		r.Get("/api/restaurants/menu_items/restaurant/{id}", restaurantsProxyHandler.ServeHTTP)

		r.Get("/api/restaurants/health", restaurantsProxyHandler.ServeHTTP)
		r.Get("/api/orders/health", ordersProxyHandler.ServeHTTP)
		r.Get("/api/couriers/health", couriersProxyHandler.ServeHTTP)
	})

	r.Group(func(r chi.Router) {
		slog.Info("setting up protected routes")
		r.Use(middleware.AuthMiddleware)

		r.Get("/api/users/users", usersProxyHandler.ServeHTTP)

		r.Post("/api/restaurants/restaurants", restaurantsProxyHandler.ServeHTTP)
		r.Put("/api/restaurants/restaurants/{id}", restaurantsProxyHandler.ServeHTTP)
		r.Delete("/api/restaurants/restaurants/{id}", restaurantsProxyHandler.ServeHTTP)

		r.Get("/api/restaurants/menu_items", restaurantsProxyHandler.ServeHTTP)
		r.Post("/api/restaurants/menu_items", restaurantsProxyHandler.ServeHTTP)
		r.Put("/api/restaurants/menu_items/{id}", restaurantsProxyHandler.ServeHTTP)
		r.Delete("/api/restaurants/menu_items/{id}", restaurantsProxyHandler.ServeHTTP)

		r.Mount("/api/orders", http.StripPrefix("/api/orders", ordersProxy))
		r.Mount("/api/couriers", http.StripPrefix("/api/couriers", couriersProxy))
	})

	httpServer := &http.Server{
		Addr:    ":" + config.Cfg.HTTP.Port,
		Handler: r,
	}

	go func() {
		slog.Info("starting API Gateway", "port", config.Cfg.HTTP.Port)
		if err := httpServer.ListenAndServe(); err != nil {
			slog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	slog.Info("shutting down servers...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("error shutting down servers", "error", err)
	}
	slog.Info("HTTP server stopped.")

	slog.Info("service gracefully stopped.")
}

func createReverseProxy(targetURL string) *httputil.ReverseProxy {
	remote, err := url.Parse(targetURL)
	if err != nil {
		slog.Error("failed to parse target url", "url", targetURL, "error", err)
		os.Exit(1)
	}

	return httputil.NewSingleHostReverseProxy(remote)
}

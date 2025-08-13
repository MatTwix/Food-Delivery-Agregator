package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/api-gateway/config"
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

	restaurantsServiceUrl, err := url.Parse(config.Cfg.URLs.RestaurantsService)
	if err != nil {
		slog.Error("failed to parse RESTAURANTS_SERVICE_URL", "error", err)
		os.Exit(1)
	}
	ordersServiceUrl, err := url.Parse(config.Cfg.URLs.OrdersService)
	if err != nil {
		slog.Error("failed to parse ORDERS_SERVICE_URL", "error", err)
		os.Exit(1)
	}
	couriersServiceUrl, err := url.Parse(config.Cfg.URLs.CouriersService)
	if err != nil {
		slog.Error("failed to parse COURIERS_SERVICE_URL", "error", err)
		os.Exit(1)
	}
	usersServiceUrl, err := url.Parse(config.Cfg.URLs.UsersService)
	if err != nil {
		slog.Error("failed to parse USERS_SERVICE_URL", "error", err)
		os.Exit(1)
	}

	restaurantsProxy := httputil.NewSingleHostReverseProxy(restaurantsServiceUrl)
	oredersProxy := httputil.NewSingleHostReverseProxy(ordersServiceUrl)
	couriersProxy := httputil.NewSingleHostReverseProxy(couriersServiceUrl)
	usersProxy := httputil.NewSingleHostReverseProxy(usersServiceUrl)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		slog.Info("incoming request", "path", path)

		if after, ok := strings.CutPrefix(path, "/api/restaurants"); ok {
			r.URL.Path = after
			slog.Info("forwarding to restaurants-service", "path", r.URL.Path)
			restaurantsProxy.ServeHTTP(w, r)
			return
		}

		if after, ok := strings.CutPrefix(path, "/api/orders"); ok {
			r.URL.Path = after
			slog.Info("forwarding to orders-service", "path", r.URL.Path)
			oredersProxy.ServeHTTP(w, r)
			return
		}

		if after, ok := strings.CutPrefix(path, "/api/couriers"); ok {
			r.URL.Path = after
			slog.Info("forwarding to couriers-service", "path", r.URL.Path)
			couriersProxy.ServeHTTP(w, r)
			return
		}

		if after, ok := strings.CutPrefix(path, "/api/users"); ok {
			r.URL.Path = after
			slog.Info("forwarding to users-service", "path", r.URL.Path)
			usersProxy.ServeHTTP(w, r)
			return
		}

		http.Error(w, "Not found", http.StatusNotFound)
	})

	httpServer := &http.Server{
		Addr:    ":" + config.Cfg.HTTP.Port,
		Handler: mux,
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

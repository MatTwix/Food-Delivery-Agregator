package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

	if config.Cfg.URLs.RestaurantsService == "" {
		log.Fatal("RESTAURANTS_SERVICE_URL is not set")
	}
	if config.Cfg.URLs.OrdersService == "" {
		log.Fatal("ORDERS_SERVICE_URL is not set")
	}
	if config.Cfg.URLs.CouriersService == "" {
		log.Fatal("COURIERS_SERVICE_URL is not set")
	}

	restaurantsServiceUrl, err := url.Parse(config.Cfg.URLs.RestaurantsService)
	if err != nil {
		log.Fatalf("Error parsing RESTAURANTS_SERVICE_URL: %v", err)
	}
	ordersServiceUrl, err := url.Parse(config.Cfg.URLs.OrdersService)
	if err != nil {
		log.Fatalf("Error parsing ORDERS_SERVICE_URL: %v", err)
	}
	couriersServiceUrl, err := url.Parse(config.Cfg.URLs.CouriersService)
	if err != nil {
		log.Fatalf("Error parsing COURIERS_SERVICE_URL: %v", err)
	}

	restaurantsProxy := httputil.NewSingleHostReverseProxy(restaurantsServiceUrl)
	oredersProxy := httputil.NewSingleHostReverseProxy(ordersServiceUrl)
	couriersProxy := httputil.NewSingleHostReverseProxy(couriersServiceUrl)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		log.Printf("Incoming request for: %s", path)

		if after, ok := strings.CutPrefix(path, "/api/restaurants"); ok {
			r.URL.Path = after
			log.Printf("Forwarding to restaurants-service with path: %s", r.URL.Path)
			restaurantsProxy.ServeHTTP(w, r)
			return
		}

		if after, ok := strings.CutPrefix(path, "/api/orders"); ok {
			r.URL.Path = after
			log.Printf("Forwarding to orders-service with path: %s", r.URL.Path)
			oredersProxy.ServeHTTP(w, r)
			return
		}

		if after, ok := strings.CutPrefix(path, "/api/couriers"); ok {
			r.URL.Path = after
			log.Printf("Forwarding to couriers-service with path: %s", r.URL.Path)
			couriersProxy.ServeHTTP(w, r)
			return
		}

		http.Error(w, "Not found", http.StatusNotFound)
	})

	httpServer := &http.Server{
		Addr:    ":" + config.Cfg.HTTP.Port,
		Handler: mux,
	}

	go func() {
		log.Printf("Starting API Gateway on port %s", config.Cfg.HTTP.Port)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down servers...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down servers: %v", err)
	}
	log.Println("HTTP server stopped.")

	log.Println("Service gracefully stopped.")
}

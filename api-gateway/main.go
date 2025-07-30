package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/MatTwix/Food-Delivery-Agregator/api-gateway/config"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.RestaurantsServiceUrl == "" {
		log.Fatal("RESTAURANTS_SERVICE_URL is not set")
	}

	target, err := url.Parse(cfg.RestaurantsServiceUrl)
	if err != nil {
		log.Fatalf("Error parsing target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	http.HandleFunc("/", proxy.ServeHTTP)

	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Error staring server: %v", err)
	}
}

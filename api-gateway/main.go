package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	restaurantsServiceUrl := os.Getenv("RESTAURANTS_SERVICE_URL")
	if restaurantsServiceUrl == "" {
		log.Fatal("RESTAURANTS_SERVICE_URL is not set")
	}

	target, err := url.Parse(restaurantsServiceUrl)
	if err != nil {
		log.Fatalf("Error parsing target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	http.HandleFunc("/", proxy.ServeHTTP)

	port := ":3000"

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error staring server: %v", err)
	}
}

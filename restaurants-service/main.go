package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Restaurants service is up and running!")
	})

	port := ":3001"
	log.Printf("Starting restaurants service on port %s", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}
}

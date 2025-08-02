package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/messaging"
)

func main() {
	cfg := config.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messaging.StartConsumer(ctx)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Orders service is up and running!")
	})

	log.Printf("Starting orders service on port %s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/api"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/database"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/store"
	"google.golang.org/grpc"
)

func main() {
	config.InitConfig()
	config.InitValidator()

	messaging.InitTopicsNames()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database.NewConnection()
	db := database.DB

	messaging.InitTopics()

	restaurantStore := store.NewRestaurantsStore(db)
	menuItemStore := store.NewMenuItemStore(db)

	kafkaProducer, err := messaging.NewProducer()
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRestaurantServiceServer(grpcServer, api.NewGrpcServer(store.NewMenuItemStore(db)))

	router := api.SetupRoutes(restaurantStore, menuItemStore, kafkaProducer)
	httpServer := &http.Server{
		Addr:    ":" + config.Cfg.HTTP.Port,
		Handler: router,
	}

	go func() {
		lis, err := net.Listen("tcp", ":"+config.Cfg.GRPC.Port)
		if err != nil {
			log.Fatalf("Error listening for gPRC: %v", err)
		}

		log.Printf("gRPC server listening on port: %v", config.Cfg.GRPC.Port)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Error serving gRPC: %v", err)
		}
	}()

	go func() {
		log.Printf("Starting restaurants service on port %s", config.Cfg.HTTP.Port)

		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start service: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down servers...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	log.Println("gRPC server stopped.")

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down servers: %v", err)
	}
	log.Println("HTTP server stopped.")

	kafkaProducer.Close()
	log.Println("Kafka producer closed.")

	database.DB.Close()
	log.Println("Database connection closed.")

	log.Println("Service gracefully stopped.")
}

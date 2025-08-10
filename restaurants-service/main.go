package main

import (
	"log"
	"net"
	"net/http"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/api"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/database"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/store"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadConfig()

	config.InitValidator()

	database.NewConnection()
	defer database.DB.Close()

	messaging.InitTopics()

	db := database.DB

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GrpcPort)
		if err != nil {
			log.Fatalf("Error listening for gPRC: %v", err)
		}

		grpcServer := grpc.NewServer()
		menuItemStore := api.NewGrpcServer(store.NewMenuItemStore(db))
		pb.RegisterRestaurantServiceServer(grpcServer, menuItemStore)

		log.Printf("gRPC server listening on port: %v", cfg.GrpcPort)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Error serving gRPC: %v", err)
		}
	}()

	restaurantStore := store.NewRestaurantsStore(db)
	menuItemStore := store.NewMenuItemStore(db)

	kafkaProducer, err := messaging.NewProducer()
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}

	r := api.SetupRoutes(restaurantStore, menuItemStore, kafkaProducer)

	log.Printf("Starting restaurants service on port %s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}

	//TODO: add gracefull shutdown
}

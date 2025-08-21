package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"google.golang.org/grpc"

	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/api"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/clients"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/database"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/store"
)

func main() {
	config.InitConfig()
	config.InitValidator()
	config.InitLogger()

	messaging.InitTopicsNames()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database.NewConnection()
	db := database.DB

	messaging.InitTopics()

	restaurantStore := store.NewRestaurantStore(db)
	orderStore := store.NewOrderStore(db)

	kafkaProducer, err := messaging.NewProducer()
	if err != nil {
		slog.Error("failed to create Kafka producer", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, api.NewOrderGRPCServer(orderStore))

	restaurantGRPCClient := clients.NewResraurantServiceClient()

	router := api.SetupRoutes(restaurantStore, orderStore, restaurantGRPCClient, kafkaProducer)
	httpServer := &http.Server{
		Addr:    ":" + config.Cfg.HTTP.Port,
		Handler: router,
	}

	messaging.StartConsumers(ctx, restaurantStore, orderStore, kafkaProducer)

	go func() {
		lis, err := net.Listen("tcp", ":"+config.Cfg.GRPC.Port)
		if err != nil {
			slog.Error("failed to listen for gRPC", "error", err)
			os.Exit(1)
		}

		slog.Info("gRPC server listening", "port", config.Cfg.GRPC.Port)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("failed to serve gRPC", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		slog.Info("starting orders service", "port", config.Cfg.HTTP.Port)
		if err := httpServer.ListenAndServe(); err != nil {
			slog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	slog.Info("shutting down servers")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shut down servers", "error", err)
	}
	slog.Info("HTTP server stopped")

	kafkaProducer.Close()
	slog.Info("Kafka producer closed")

	database.DB.Close()
	slog.Info("database connection closed")

	slog.Info("service gracefully stopped")
}

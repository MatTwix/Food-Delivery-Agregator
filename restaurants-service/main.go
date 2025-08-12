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
	config.InitLogger()

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
		slog.Error("failed to create Kafka producer", "error", err)
		os.Exit(1)
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
			slog.Error("failed to listen for gPRC", "error", err)
			os.Exit(1)
		}

		slog.Info("gRPC server listening", "port", config.Cfg.GRPC.Port)

		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("failed to serve gRPC", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		slog.Info("starting restaurants service", "port", config.Cfg.HTTP.Port)

		if err := httpServer.ListenAndServe(); err != nil {
			slog.Error("failed to start service", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	slog.Info("shutting down servers")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	slog.Info("gRPC server stopped")

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

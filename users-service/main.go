package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/users-service/api"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/database"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/store"
)

func main() {
	config.InitConfig()
	config.InitValidator()
	config.InitLogger()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database.NewConnection()
	db := database.DB

	userStore := store.NewUserStore(db)

	router := api.SetupRoutes(userStore)
	httpServer := &http.Server{
		Addr:    ":" + config.Cfg.HTTP.Port,
		Handler: router,
	}

	go func() {
		slog.Info("starting users service", "port", config.Cfg.HTTP.Port)
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

	database.DB.Close()
	slog.Info("database connection closed")

	slog.Info("service gracefully stopped")
}

package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/MatTwix/Food-Delivery-Agregator/scheduler-service/config"
)

func main() {
	config.InitConfig()
	config.InitLogger()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	slog.Info("shutting down servers")

	slog.Info("service gracefully stopped")
}

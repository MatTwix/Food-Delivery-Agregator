package database

import (
	"context"
	"log/slog"
	"os"

	"github.com/MatTwix/Food-Delivery-Agregator/users-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/migrations"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/seeders"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func NewConnection() {
	if config.Cfg.DB.Source == "" {
		slog.Error("DB_SOURCE environment variable is not set")
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), config.Cfg.DB.Source)
	if err != nil {
		slog.Error("failed to create connection pool", "error", err)
		os.Exit(1)
	}

	if err := pool.Ping(context.Background()); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	DB = pool
	slog.Info("successfully connected to the database")

	migrations.Migrate(DB)
	seeders.Seed(DB)
}

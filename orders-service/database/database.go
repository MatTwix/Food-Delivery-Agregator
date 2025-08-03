package database

import (
	"context"
	"log"

	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func NewConnection() {
	cfg := config.LoadConfig()

	if cfg.DbSource == "" {
		log.Fatal("DB_SOURCE environment variable is not set")
	}

	pool, err := pgxpool.New(context.Background(), cfg.DbSource)
	if err != nil {
		log.Fatalf("Error creating connection pool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	DB = pool
	log.Println("Successfully connected to the database")

	migrations.Migrate(DB)
}

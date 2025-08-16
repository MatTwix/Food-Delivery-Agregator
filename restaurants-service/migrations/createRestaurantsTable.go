package migrations

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateRestaurantsTable(db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		slog.Error("failed to start transaction", "error", err)
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	var tableExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'restaurants');").
		Scan(&tableExists)

	if err != nil {
		slog.Error("failed to check restaurants table existance", "error", err)
		os.Exit(1)
	}

	if !tableExists {
		_, err = tx.Exec(ctx, `
			CREATE TABLE restaurants (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				owner_id UUID NOT NULL,
				name VARCHAR(255) NOT NULL,
				address TEXT NOT NULL,
				phone_number VARCHAR(50),
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);
		`)
		if err != nil {
			slog.Error("failed to create restaurants table", "error", err)
			os.Exit(1)
		}

		err = tx.Commit(ctx)
		if err != nil {
			slog.Error("failed to commit transaction", "error", err)
			os.Exit(1)
		}

		slog.Info("restaurants table created successfully")
	} else {
		tx.Rollback(ctx)
	}
}

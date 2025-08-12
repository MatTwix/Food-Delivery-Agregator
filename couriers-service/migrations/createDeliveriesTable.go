package migrations

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateDeliveriesTable(db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		slog.Error("failed to start transaction", "error", err)
	}
	defer tx.Rollback(ctx)

	var tableExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'deliveries');").
		Scan(&tableExists)
	if err != nil {
		slog.Error("failed to check deliveries table existance", "error", err)
		os.Exit(1)
	}

	if !tableExists {
		_, err = tx.Exec(ctx, `
			CREATE TABLE deliveries (
				order_id UUID PRIMARY KEY,
				courier_id UUID NOT NULL REFERENCES couriers(id),
				assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);

			CREATE INDEX IF NOT EXISTS idx_order_id ON deliveries(order_id);
		`)

		if err != nil {
			slog.Error("failed to creat deliveries table", "error", err)
			os.Exit(1)
		}

		err = tx.Commit(ctx)
		if err != nil {
			slog.Error("failed to commit transaction", "error", err)
			os.Exit(1)
		}

		slog.Info("deliveries table created successfully")
	} else {
		tx.Rollback(ctx)
	}
}

package migrations

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateMenuItemsTable(db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		slog.Error("failed to start transaction", "error", err)
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	var tableExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'menu_items');").
		Scan(&tableExists)

	if err != nil {
		slog.Error("failed to check menu_items table existance", "error", err)
	}

	if !tableExists {
		_, err = tx.Exec(ctx, `
			CREATE TABLE menu_items (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				restaurant_id UUID NOT NULL,
				name VARCHAR(255) NOT NULL,
				description TEXT,
				price NUMERIC(10, 2) NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				CONSTRAINT fk_restaurant FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE
			);

			CREATE INDEX IF NOT EXISTS idx_menu_items_restairant_id ON menu_items(restaurant_id);
		`)

		if err != nil {
			slog.Error("failed to create menu_items table", "error", err)
			os.Exit(1)
		}

		err = tx.Commit(ctx)
		if err != nil {
			slog.Error("failed to commit transaction", "error", err)
			os.Exit(1)
		}

		slog.Info("menu_items table created successfully!")
	} else {
		tx.Rollback(ctx)
	}
}

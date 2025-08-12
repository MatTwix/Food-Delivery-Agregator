package migrations

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateOrdersItemsTable(db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		slog.Error("failed to start transaction", "error", err)
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	var tableExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'orders_items');").
		Scan(&tableExists)

	if err != nil {
		slog.Error("failed to check orders_items table existance", "error", err)
		os.Exit(1)
	}

	if !tableExists {
		_, err = tx.Exec(ctx, `
			CREATE TABLE orders_items (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				order_id UUID NOT NULL,
				menu_item_id UUID NOT NULL,
				quantity INT NOT NULL,
				price NUMERIC(10, 2) NOT NULL,
				CONSTRAINT fk_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
			);
		`)
		if err != nil {
			slog.Error("failed to create orders_items table", "error", err)
			os.Exit(1)
		}

		err = tx.Commit(ctx)
		if err != nil {
			slog.Error("failed to commit transaction", "error", err)
			os.Exit(1)
		}

		slog.Info("orders_items table created successfully")
	} else {
		tx.Rollback(ctx)
	}
}

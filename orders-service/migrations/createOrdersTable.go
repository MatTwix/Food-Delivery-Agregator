package migrations

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateOrdersTable(db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	var tableExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'orders');").
		Scan(&tableExists)

	if err != nil {
		log.Fatalf("Error cheking orders table existance: %v", err)
	}

	if !tableExists {
		_, err = tx.Exec(ctx, `
			CREATE TABLE orders (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				restaurant_id UUID NOT NULL,
				user_id UUID,
				total_price NUMERIC(10, 2) NOT NULL,
				courier_id UUID,
				status VARCHAR(50) NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);

			CREATE INDEX IF NOT EXISTS idx_orders_restaurant_id ON orders(restaurant_id);
			CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
		`)
		if err != nil {
			log.Fatalf("Error creating orders table: %v", err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			log.Fatalf("Error commiting transaction: %v", err)
		}

		log.Println("Orders table created successfully!")
	} else {
		tx.Rollback(ctx)
	}
}

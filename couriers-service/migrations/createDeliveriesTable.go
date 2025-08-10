package migrations

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateDeliveriesTable(db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	var tableExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'deliveries');").
		Scan(&tableExists)
	if err != nil {
		log.Fatalf("Error checking deliveries table existance: %v", err)
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
			log.Fatalf("Error creating deliveries table: %v", err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			log.Fatalf("Error commiting transaction: %v", err)
		}

		log.Println("Deliveries table created successfully!")
	} else {
		tx.Rollback(ctx)
	}
}

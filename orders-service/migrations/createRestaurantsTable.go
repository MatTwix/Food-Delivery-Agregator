package migrations

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateRestaurantsTable(db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	var tableExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'restaurants');").
		Scan(&tableExists)

	if err != nil {
		log.Fatalf("Error cheking restaurants table existance: %v", err)
	}

	if !tableExists {
		_, err = tx.Exec(ctx, `
			CREATE TABLE restaurants (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(255) NOT NULL,
				is_active BOOLEAN NOT NULL DEFAULT true,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);
		`)
		if err != nil {
			log.Fatalf("Error creating restaurants table: %v", err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			log.Fatalf("Error commiting transaction: %v", err)
		}

		log.Println("Restaurants table created successfully!")
	} else {
		tx.Rollback(ctx)
	}
}

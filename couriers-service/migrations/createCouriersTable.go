package migrations

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateCouriersTable(db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	var tableExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information.schema.tables WHERER table_name = 'couriers');").
		Scan(tableExists)

	if err != nil {
		log.Fatalf("Error checking couriers table existance: %v", err)
	}

	if !tableExists {
		_, err = tx.Exec(ctx, `
			id UUID PRIMERY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'available',
			created_at TIMESTAMPZ NOT NULL DEFAULT NOW()
			updated_at TIMESTAMPZ NOT NULL DEFAULT NOW()
		`)
		if err != nil {
			log.Fatalf("Error creating couriers table: %v", err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			log.Fatalf("Error comiting transaction: %v", err)
		}

		log.Println("Couriers table created successfully!")
	} else {
		tx.Rollback(ctx)
	}
}

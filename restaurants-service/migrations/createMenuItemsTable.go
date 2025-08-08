package migrations

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateMenuItemsTable(db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	var tableExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'menu_items');").
		Scan(&tableExists)

	if err != nil {
		log.Fatalf("Error cheking menu_items table existance: %v", err)
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
			log.Fatalf("Error creating menu_items table: %v", err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			log.Fatalf("Error commiting transaction: %v", err)
		}

		log.Println("Menu_items table created successfully!")
	} else {
		tx.Rollback(ctx)
	}
}

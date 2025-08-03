package migrations

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateOrdersItemsTable(db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	var tableExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'orders_items');").
		Scan(&tableExists)

	if err != nil {
		log.Fatalf("Error cheking orders_items table existance: %v", err)
	}

	if !tableExists {
		_, err = tx.Exec(ctx, `
			CREATE TABLE orders_items (
				id SERIAL PRIMARY KEY,
				order_id UUID NOT NULL,
				menu_item_id UUID NOT NULL,
				quantity INT NOT NULL,
				price NUMERIC(10, 2) NOT NULL,
				CONSTRAINT fk_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
			);
		`)
		if err != nil {
			log.Fatalf("Error creating orders_items table: %v", err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			log.Fatalf("Error commiting transaction: %v", err)
		}

		log.Println("Orders_items table created successfully!")
	} else {
		tx.Rollback(ctx)
	}
}

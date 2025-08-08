package store

import (
	"context"
	"log"

	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderStore struct {
	db *pgxpool.Pool
}

func NewOrderStore(db *pgxpool.Pool) *OrderStore {
	return &OrderStore{db: db}
}

func (s *OrderStore) Create(ctx context.Context, order *models.Order) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	orderQuery := `
		INSERT INTO orders (restaurant_id, total_price, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRow(ctx, orderQuery, order.RestaurantID, order.TotalPrice, order.Status).
		Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return err
	}

	itemRows := [][]any{}
	for _, item := range order.Items {
		itemRows = append(itemRows, []any{order.ID, item.MenuItemID, item.Quantity, item.Price})
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"orders_items"},
		[]string{"order_id", "menu_item_id", "quantity", "price"},
		pgx.CopyFromRows(itemRows),
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *OrderStore) UpdateStatus(ctx context.Context, orderID string, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := s.db.Exec(ctx, query, status, orderID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		log.Printf("Warning: order with ID %s is not found for status update.", orderID)
	}

	return nil
}

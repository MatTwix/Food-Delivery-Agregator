package store

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

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

func (s *OrderStore) GetAll(ctx context.Context) ([]models.Order, error) {
	orderQuery := `
		SELECT
		id, restaurant_id, user_id, total_price, status, courier_id, retry_count, max_retry_count, next_retry_at, created_at, updated_at
		FROM
		orders
	`

	var orders []models.Order

	orderRows, err := s.db.Query(ctx, orderQuery)
	if err != nil {
		return orders, nil
	}

	for orderRows.Next() {
		var order models.Order

		err := orderRows.Scan(
			&order.ID,
			&order.RestaurantID,
			&order.UserID,
			&order.TotalPrice,
			&order.Status,
			&order.CourierID,
			&order.RetryCount,
			&order.MaxRetryCount,
			&order.NextRetryAt,
			&order.CreatedAt,
			&order.UpdatedAt,
		)

		if err != nil {
			return orders, err
		}

		itemsQuery := `
			SELECT 
			id, order_id, menu_item_id, quantity, price
			FROM
			orders_items
			WHERE
			order_id = $1
		`

		itemRows, err := s.db.Query(ctx, itemsQuery, order.ID)
		if err != nil {
			return orders, err
		}

		for itemRows.Next() {
			var item models.OrderItem
			if err := itemRows.Scan(&item.ID, &item.OrderID, &item.MenuItemID, &item.Quantity, &item.Price); err != nil {
				return orders, err
			}

			order.Items = append(order.Items, item)
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (s *OrderStore) GetByID(ctx context.Context, id string) (models.Order, error) {
	orderQuery := `
		SELECT
		id, restaurant_id, user_id, total_price, status, courier_id, retry_count, max_retry_count, next_retry_at, created_at, updated_at
		FROM
		orders
		WHERE id = $1
	`
	var order models.Order

	err := s.db.QueryRow(ctx, orderQuery, id).Scan(
		&order.ID,
		&order.RestaurantID,
		&order.UserID,
		&order.TotalPrice,
		&order.Status,
		&order.CourierID,
		&order.RetryCount,
		&order.MaxRetryCount,
		&order.NextRetryAt,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		return order, err
	}

	itemsQuery := `
		SELECT 
		id, order_id, menu_item_id, quantity, price
		FROM
		orders_items
		WHERE
		order_id = $1
	`

	rows, err := s.db.Query(ctx, itemsQuery, order.ID)
	if err != nil {
		return order, err
	}

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.MenuItemID, &item.Quantity, &item.Price); err != nil {
			return order, err
		}

		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (s *OrderStore) GetOwnerID(ctx context.Context, orderID string) (string, error) {
	query := `
		SELECT user_id
		FROM orders
		WHERE id = $1
	`

	var ownerID string

	err := s.db.QueryRow(ctx, query, orderID).Scan(&ownerID)

	return ownerID, err
}

func (s *OrderStore) GetStatus(ctx context.Context, orderID string) (string, error) {
	query := `
		SELECT status
		FROM orders
		WHERE id = $1
	`

	var status string

	err := s.db.QueryRow(ctx, query, orderID).Scan(&status)

	return status, err
}

func (s *OrderStore) GetForRetry(ctx context.Context, status string, nextRetryAtLte int64, limit int32) ([]models.Order, error) {
	orderQuery := `
		SELECT
		id, restaurant_id, user_id, total_price, status, courier_id, retry_count, max_retry_count, next_retry_at, created_at, updated_at
		FROM orders
		WHERE status = $1 AND next_retry_at <= $2 AND retry_count < max_retry_count
		ORDER BY next_retry_at ASC
	`

	args := []any{status}

	if nextRetryAtLte == 0 {
		orderQuery = strings.Replace(orderQuery, "$2", "NOW()", 1)
	} else {
		args = append(args, time.Unix(nextRetryAtLte, 0))
	}

	if limit > 0 {
		orderQuery += " LIMIT $3"
		args = append(args, limit)
	}

	var orders []models.Order

	orderRows, err := s.db.Query(ctx, orderQuery, args...)
	if err != nil {
		return orders, nil
	}

	for orderRows.Next() {
		var order models.Order

		err := orderRows.Scan(
			&order.ID,
			&order.RestaurantID,
			&order.UserID,
			&order.TotalPrice,
			&order.Status,
			&order.CourierID,
			&order.RetryCount,
			&order.MaxRetryCount,
			&order.NextRetryAt,
			&order.CreatedAt,
			&order.UpdatedAt,
		)

		if err != nil {
			return orders, err
		}

		itemsQuery := `
			SELECT 
			id, order_id, menu_item_id, quantity, price
			FROM
			orders_items
			WHERE
			order_id = $1
		`

		itemRows, err := s.db.Query(ctx, itemsQuery, order.ID)
		if err != nil {
			return orders, err
		}

		for itemRows.Next() {
			var item models.OrderItem
			if err := itemRows.Scan(&item.ID, &item.OrderID, &item.MenuItemID, &item.Quantity, &item.Price); err != nil {
				return orders, err
			}

			order.Items = append(order.Items, item)
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (s *OrderStore) GetTotalPrice(ctx context.Context, orderID string) (float64, error) {
	query := `
		SELECT total_price
		FROM orders
		WHERE id = $1
	`

	var totalPrice float64
	err := s.db.QueryRow(ctx, query, orderID).Scan(&totalPrice)

	return totalPrice, err
}

func (s *OrderStore) Create(ctx context.Context, order *models.Order) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	orderQuery := `
		INSERT INTO orders (user_id, restaurant_id, total_price, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRow(ctx, orderQuery, order.UserID, order.RestaurantID, order.TotalPrice, order.Status).
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

func (s *OrderStore) IncreaseRetryCounter(ctx context.Context, orderID string) (retriesExceeded bool, err error) {
	query := `
		UPDATE orders
		SET retry_count = retry_count + 1
		WHERE id = $1 AND retry_count < max_retry_count
		RETURNING retry_count, max_retry_count
	`

	var retryCount, maxRetryCount int

	if err = s.db.QueryRow(ctx, query, orderID).Scan(&retryCount, &maxRetryCount); err != nil {
		if err == pgx.ErrNoRows {
			return true, nil
		}

		return false, err
	}

	return false, nil
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
		slog.Warn("order is not found for status update", "order_id", orderID)
	}

	return nil
}

func (s *OrderStore) AssignCourier(ctx context.Context, orderID, courierID string) error {
	query := `
		UPDATE orders
		SET status = 'awaiting_pickup', courier_id = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := s.db.Exec(ctx, query, courierID, orderID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("Error assigning courier: cannot found order with id: " + orderID)
	}

	return nil
}

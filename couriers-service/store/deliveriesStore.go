package store

import (
	"context"

	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeliveryStore struct {
	db *pgxpool.Pool
}

func NewDeliveryStore(db *pgxpool.Pool) *DeliveryStore {
	return &DeliveryStore{db: db}
}

func (s *DeliveryStore) GetCourierIDByOrderID(ctx context.Context, orderID string) (string, error) {
	query := `
		SELECT
		courier_id
		FROM
		deliveries
		WHERE order_id = $1
	`

	var courierID string

	err := s.db.QueryRow(ctx, query, orderID).Scan(&courierID)

	return courierID, err
}

func (s *DeliveryStore) Create(ctx context.Context, delivery *models.Delivery) error {
	query := `
		INSERT INTO 
		deliveries
		(order_id, courier_id)
		VALUES
		($1, $2)
		RETURNING assigned_at
	`

	err := s.db.QueryRow(ctx, query, delivery.OrderID, delivery.CourierID).
		Scan(&delivery.AssignedAt)

	return err
}

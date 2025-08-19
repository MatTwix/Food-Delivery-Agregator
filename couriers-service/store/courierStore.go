package store

import (
	"context"
	"errors"

	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourierStore struct {
	db *pgxpool.Pool
}

func NewCourierStore(db *pgxpool.Pool) *CourierStore {
	return &CourierStore{db: db}
}

func (s *CourierStore) GetAll(ctx context.Context) ([]models.Courier, error) {
	query := `
		SELECT
		id, name, status, created_at, updated_at
		FROM
		couriers
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var couriers []models.Courier
	for rows.Next() {
		var courier models.Courier
		if err := rows.Scan(
			&courier.ID,
			&courier.Name,
			&courier.Status,
			&courier.CreatedAt,
			&courier.UpdatedAt,
		); err != nil {
			return nil, err
		}
		couriers = append(couriers, courier)
	}

	return couriers, nil
}

func (s *CourierStore) GetAvailable(ctx context.Context) (models.Courier, error) {
	query := `
		SELECT
		id, name, status, created_at, updated_at
		FROM
		couriers
		WHERE
		status = 'available'
	`
	var courier models.Courier

	err := s.db.QueryRow(ctx, query).Scan(
		&courier.ID,
		&courier.Name,
		&courier.Status,
		&courier.CreatedAt,
		&courier.UpdatedAt,
	)

	return courier, err
}

func (s *CourierStore) Create(ctx context.Context, courier *models.Courier) error {
	query := `
		INSERT INTO couriers
		(id, name)
		VALUES
		($1, $2)
		RETURNING id, status, created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query, courier.ID, courier.Name).
		Scan(&courier.ID, &courier.Status, &courier.CreatedAt, &courier.UpdatedAt)

	return err
}

func (s *CourierStore) Update(ctx context.Context, courier *models.Courier) error {
	query := `
		UPDATE couriers
		SET name = $1
		WHERE id = $2
		RETURNING created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query, courier.Name, courier.ID).
		Scan(&courier.CreatedAt, &courier.UpdatedAt)

	return err
}

func (s *CourierStore) UpdateStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE couriers
		SET status = $1
		WHERE id = $2
	`

	_, err := s.db.Exec(ctx, query, status, id)

	return err
}

func (s *CourierStore) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM couriers WHERE id = $1
	`

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("courier not found")
	}

	return nil
}

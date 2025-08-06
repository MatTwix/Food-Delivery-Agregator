package store

import (
	"context"

	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RestaurantStore struct {
	db *pgxpool.Pool
}

func NewRestaurantStore(db *pgxpool.Pool) *RestaurantStore {
	return &RestaurantStore{db: db}
}

func (s *RestaurantStore) Upsert(ctx context.Context, restaurant *models.Restaurant) error {
	query := `
		INSERT INTO restaurants (id, name, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at;
	`

	_, err := s.db.Exec(ctx, query, restaurant.ID, restaurant.Name, restaurant.UpdatedAt)

	return err
}

func (s *RestaurantStore) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM 
		restaurants
		WHERE
		id = $1
	`

	_, err := s.db.Exec(ctx, query, id)

	return err
}

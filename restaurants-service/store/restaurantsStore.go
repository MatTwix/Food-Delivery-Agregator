package store

import (
	"context"

	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RestaurantStore struct {
	db *pgxpool.Pool
}

func NewRestaurantsStore(db *pgxpool.Pool) *RestaurantStore {
	return &RestaurantStore{db: db}
}

func (s *RestaurantStore) Create(ctx context.Context, restaurant *models.Restaurant) error {
	query := `
		INSERT INTO restaurants 
		(name, address, phone_number)
		VALUES
		($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, restaurant.Name, restaurant.Address, restaurant.PhoneNumber).
		Scan(&restaurant.ID, &restaurant.CreatedAt, &restaurant.UpdatedAt)

	return err
}

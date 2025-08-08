package store

import (
	"context"
	"errors"

	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RestaurantStore struct {
	db *pgxpool.Pool
}

func NewRestaurantsStore(db *pgxpool.Pool) *RestaurantStore {
	return &RestaurantStore{db: db}
}

func (s *RestaurantStore) GetAll(ctx context.Context) ([]models.Restaurant, error) {
	query := `
		SELECT 
		id, name, address, phone_number, created_at, updated_at 
		FROM
		restaurants
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var restaurants []models.Restaurant
	for rows.Next() {
		var restaurant models.Restaurant
		if err := rows.Scan(
			&restaurant.ID,
			&restaurant.Name,
			&restaurant.Address,
			&restaurant.PhoneNumber,
			&restaurant.CreatedAt,
			&restaurant.UpdatedAt,
		); err != nil {
			return nil, err
		}
		restaurants = append(restaurants, restaurant)
	}

	return restaurants, nil
}

func (s *RestaurantStore) GetByID(ctx context.Context, id string) (models.Restaurant, error) {
	query := `
		SELECT
		name, address, phone_number, created_at, updated_at
		FROM
		restaurants
		WHERE
		id = $1
	`

	restaurant := models.Restaurant{
		ID: id,
	}

	err := s.db.QueryRow(ctx, query, id).
		Scan(
			&restaurant.Name,
			&restaurant.Address,
			&restaurant.PhoneNumber,
			&restaurant.CreatedAt,
			&restaurant.UpdatedAt,
		)

	return restaurant, err
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

func (s *RestaurantStore) Update(ctx context.Context, restaurant *models.Restaurant) error {
	query := `
		UPDATE restaurants
		SET
		name = $1, address = $2, phone_number = $3
		WHERE
		id = $4
		RETURNING created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query, restaurant.Name, restaurant.Address, restaurant.PhoneNumber, restaurant.ID).
		Scan(&restaurant.CreatedAt, &restaurant.UpdatedAt)

	return err
}

func (s *RestaurantStore) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM restaurants
		WHERE 
		id = $1
	`

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("restaurant not found")
	}

	return nil
}

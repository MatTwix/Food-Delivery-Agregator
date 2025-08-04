package store

import (
	"context"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
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

func (s *RestaurantStore) GetMenuItemsByIDs(ctx context.Context, itemIDs []string) ([]*pb.MenuItem, error) {
	query := "SELECT id, name, price FROM menu_items WHERE id = ANY($1)"
	rows, err := s.db.Query(ctx, query, itemIDs)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*pb.MenuItem
	for rows.Next() {
		var item pb.MenuItem
		var price float64
		if err := rows.Scan(&item.Id, &item.Name, &price); err != nil {
			return nil, err
		}
		item.Price = price
		items = append(items, &item)
	}

	return items, nil
}

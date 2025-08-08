package store

import (
	"context"
	"errors"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MenuItemStore struct {
	db *pgxpool.Pool
}

func NewMenuItemStore(db *pgxpool.Pool) *MenuItemStore {
	return &MenuItemStore{db: db}
}

func (s *MenuItemStore) GetAll(ctx context.Context) ([]models.MenuItem, error) {
	query := `
		SELECT
		id, restaurant_id, name, description, price, created_at, updated_at
		FROM
		menu_items
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menuItems []models.MenuItem

	for rows.Next() {
		var menuItem models.MenuItem
		if err := rows.Scan(
			&menuItem.ID,
			&menuItem.RestaurantID,
			&menuItem.Name,
			&menuItem.Desctiption,
			&menuItem.Price,
			&menuItem.CreatedAt,
			&menuItem.UpdatedAt,
		); err != nil {
			return nil, err
		}
		menuItems = append(menuItems, menuItem)
	}

	return menuItems, nil
}

func (s *MenuItemStore) GetMenuItemsByIDs(ctx context.Context, itemIDs []string) ([]*pb.MenuItem, error) {
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

func (s *MenuItemStore) Create(ctx context.Context, menuItem *models.MenuItem) error {
	query := `
		INSERT INTO menu_items
		(restauarnt_id, name, description, price)
		VALUES
		($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query, menuItem.RestaurantID, menuItem.Name, menuItem.Desctiption, menuItem.Price).
		Scan(&menuItem.ID, &menuItem.CreatedAt, &menuItem.UpdatedAt)

	return err
}

func (s *MenuItemStore) Update(ctx context.Context, menuItem *models.MenuItem) error {
	query := `
		UPDATE menu_items
		SET name = $1, description = $2, price = $3)
		WHERE id = $4
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query, menuItem.RestaurantID, menuItem.Name, menuItem.Desctiption, menuItem.Price).
		Scan(&menuItem.ID, &menuItem.CreatedAt, &menuItem.UpdatedAt)

	return err
}

func (s *MenuItemStore) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM menu_items
		WHERE
		id = $1
	`

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("menu item not found")
	}

	return nil
}

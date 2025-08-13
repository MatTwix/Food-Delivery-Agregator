package store

import (
	"context"

	"github.com/MatTwix/Food-Delivery-Agregator/users-service/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStore struct {
	db *pgxpool.Pool
}

func NewUserStore(db *pgxpool.Pool) *UserStore {
	return &UserStore{
		db: db,
	}
}

func (s *UserStore) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users 
		(email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query, user.Email, user.PasswordHash, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	return err
}

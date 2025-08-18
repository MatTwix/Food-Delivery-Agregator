package store

import (
	"context"
	"errors"

	"github.com/MatTwix/Food-Delivery-Agregator/common/auth"
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

func (s *UserStore) GetAll(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
	`

	var users []models.User
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return users, err
	}

	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return users, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (s *UserStore) GetByID(ctx context.Context, id string) (models.User, error) {
	query := `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := s.db.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	return user, err
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (models.User, error) {
	query := `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := s.db.QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	return user, err
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

func (s *UserStore) ChangeRole(ctx context.Context, id string, role auth.Role) error {
	query := `
		UPDATE users
		SET role = $1
		WHERE id = $2
	`

	result, err := s.db.Exec(ctx, query, role.String(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

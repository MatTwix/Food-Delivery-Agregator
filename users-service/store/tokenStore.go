package store

import (
	"context"

	"github.com/MatTwix/Food-Delivery-Agregator/users-service/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenStore struct {
	db *pgxpool.Pool
}

func NewTokenStore(db *pgxpool.Pool) *TokenStore {
	return &TokenStore{db: db}
}

func (s *TokenStore) SaveRefreshToken(ctx context.Context, token *models.Token) error {
	query := `
		INSERT INTO refresh_tokens
		(user_id, token, expires_at)
		VALUES
		($1, $2, $3)
	`

	_, err := s.db.Exec(ctx, query, token.UserID, token.Token, token.ExpiresAt)
	return err
}

func (s *TokenStore) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token = $1
	`

	_, err := s.db.Exec(ctx, query, token)
	return err
}

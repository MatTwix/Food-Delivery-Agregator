package store

import (
	"context"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/users-service/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenStore struct {
	db *pgxpool.Pool
}

func NewTokenStore(db *pgxpool.Pool) *TokenStore {
	return &TokenStore{db: db}
}

func (s *TokenStore) GetExpiredTokensIDs(ctx context.Context, date time.Time) ([]string, error) {
	query := `
		SELECT token_id
		FROM refresh_tokens
		WHERE expires_at <= $1
	`

	var tokenIDs []string

	rows, err := s.db.Query(ctx, query, date)
	if err != nil {
		return tokenIDs, err
	}

	for rows.Next() {
		var tokenID string

		if err := rows.Scan(&tokenID); err != nil {
			return tokenIDs, err
		}

		tokenIDs = append(tokenIDs, tokenID)
	}

	return tokenIDs, nil
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

func (s *TokenStore) DeleteRefreshTokenByID(ctx context.Context, token_id string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token_id = $1
	`

	result, err := s.db.Exec(ctx, query, token_id)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (s *TokenStore) DeleteRefreshTokenByToken(ctx context.Context, token string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token = $1
	`

	result, err := s.db.Exec(ctx, query, token)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

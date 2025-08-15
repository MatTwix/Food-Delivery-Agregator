package migrations

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateRefreshTokensTable(db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		slog.Error("failed to start transaction", "error", err)
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	var tableExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'refresh_tokens');").
		Scan(&tableExists)

	if err != nil {
		slog.Error("failed to check refresh_tokens table existance", "error", err)
		os.Exit(1)
	}

	if !tableExists {
		_, err = tx.Exec(ctx, `
			CREATE TABLE refresh_tokens (
				token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				token TEXT UNIQUE NOT NULL,
				expires_at TIMESTAMPTZ NOT NULL
			);

			CREATE INDEX IF NOT EXISTS idx_refresh_user_id ON refresh_tokens(user_id);
		`)

		if err != nil {
			slog.Error("failed to create refresh_tokens table", "error", err)
			os.Exit(1)
		}

		err := tx.Commit(ctx)
		if err != nil {
			slog.Error("failed to commit transaction", "error", err)
			os.Exit(1)
		}

		slog.Info("refresh_tokens table created successfully")
	} else {
		tx.Rollback(ctx)
	}
}

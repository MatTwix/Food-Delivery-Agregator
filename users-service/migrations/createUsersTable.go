package migrations

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUsersTable(db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		slog.Error("failed to start transaction", "error", err)
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	var tableExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users');").
		Scan(&tableExists)

	if err != nil {
		slog.Error("failed to check orders table existance", "error", err)
		os.Exit(1)
	}

	if !tableExists {
		_, err = tx.Exec(ctx, `
			CREATE TABLE users (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				email VARCHAR(255) UNIQUE NOT NULL,
				password_hash TEXT NOT NULL,
				role VARCHAR(50) NOT NULL DEFAULT 'customer',
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);

			CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		`)

		if err != nil {
			slog.Error("failed to create users table", "error", err)
			os.Exit(1)
		}

		err := tx.Commit(ctx)
		if err != nil {
			slog.Error("failed to commit transaction", "error", err)
			os.Exit(1)
		}

		slog.Info("users table created successfully")
	} else {
		tx.Rollback(ctx)
	}
}

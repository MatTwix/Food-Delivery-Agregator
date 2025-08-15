package seeders

import (
	"context"
	"log/slog"
	"os"

	commonAuth "github.com/MatTwix/Food-Delivery-Agregator/common/auth"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/auth"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AddAdmin(db *pgxpool.Pool) {
	if config.Cfg.Admin.Email == "" || config.Cfg.Admin.Password == "" {
		slog.Error("admin email or password not set up")
		os.Exit(1)
	}

	passwordHash, err := auth.HashPassword(config.Cfg.Admin.Password)
	if err != nil {
		slog.Error("failed to hash admin password", "error", err)
		os.Exit(1)
	}

	usersToInsert := []models.User{
		{Email: config.Cfg.Admin.Email, PasswordHash: passwordHash, Role: commonAuth.RoleAdmin.String()},
	}

	ctx := context.Background()

	for _, user := range usersToInsert {
		tx, err := db.Begin(ctx)
		if err != nil {
			slog.Error("failed to start transaction", "error", err)
			os.Exit(1)
		}
		defer tx.Rollback(ctx)

		var userExists bool
		err = tx.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)",
			user.Email).Scan(&userExists)

		if err != nil {
			slog.Error("failed to check user existance", "error", err)
			continue
		}

		if !userExists {
			_, err := db.Exec(ctx, `
				INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)
			`, user.Email, user.PasswordHash, user.Role)

			if err != nil {
				slog.Error("failed to insert user", "error", err)
				continue
			}

			if err := tx.Commit(ctx); err != nil {
				slog.Error("failed to commit transaction", "error", err)
				continue
			}

			slog.Info("user has been seeded", "email", user.Email)
		} else {
			tx.Rollback(ctx)
		}
	}
}

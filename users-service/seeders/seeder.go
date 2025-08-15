package seeders

import "github.com/jackc/pgx/v5/pgxpool"

func Seed(db *pgxpool.Pool) {
	AddAdmin(db)
}

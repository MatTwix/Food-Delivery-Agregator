package migrations

import "github.com/jackc/pgx/v5/pgxpool"

func Migrate(db *pgxpool.Pool) {
	CreateCouriersTable(db)
	CreateDeliveriesTable(db)
}

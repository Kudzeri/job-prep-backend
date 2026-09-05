package database

import (
	"github.com/Kudzeri/job-prep-backend/pkg/database/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func EnsureSchema(pool *pgxpool.Pool) error {
	return migrations.Ensure(pool)
}

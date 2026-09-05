package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SchemaFile struct {
	Name       string
	Statements []string
}

func Ensure(pool *pgxpool.Pool) error {
	ctx := context.Background()

	for _, schemaFile := range All() {
		for _, statement := range schemaFile.Statements {
			if _, err := pool.Exec(ctx, statement); err != nil {
				return fmt.Errorf("schema %s: %w", schemaFile.Name, err)
			}
		}
	}

	return nil
}

func All() []SchemaFile {
	return []SchemaFile{
		Users(),
		EmailOTPCodes(),
		RefreshTokens(),
		TelegramAuthSessions(),
	}
}

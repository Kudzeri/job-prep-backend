package migrations

func Users() SchemaFile {
	return SchemaFile{
		Name: "users",
		Statements: []string{
			`CREATE TABLE IF NOT EXISTS users (
				id BIGSERIAL PRIMARY KEY,
				email TEXT UNIQUE,
				email_verified BOOLEAN NOT NULL DEFAULT FALSE,
				telegram_id BIGINT UNIQUE,
				telegram_username TEXT,
				first_name TEXT,
				is_onboarded BOOLEAN NOT NULL DEFAULT FALSE,
				role TEXT NOT NULL DEFAULT 'user',
				is_active BOOLEAN NOT NULL DEFAULT TRUE,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
		},
	}
}

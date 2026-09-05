package migrations

func TelegramAuthSessions() SchemaFile {
	return SchemaFile{
		Name: "telegram_auth_sessions",
		Statements: []string{
			`CREATE TABLE IF NOT EXISTS telegram_auth_sessions (
				id BIGSERIAL PRIMARY KEY,
				auth_token TEXT NOT NULL UNIQUE,
				telegram_id BIGINT UNIQUE,
				status TEXT NOT NULL DEFAULT 'pending',
				expires_at TIMESTAMPTZ NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
			`CREATE INDEX IF NOT EXISTS idx_telegram_auth_sessions_expires_at ON telegram_auth_sessions (expires_at)`,
		},
	}
}

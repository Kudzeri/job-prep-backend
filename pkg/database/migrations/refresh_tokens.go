package migrations

func RefreshTokens() SchemaFile {
	return SchemaFile{
		Name: "refresh_tokens",
		Statements: []string{
			`CREATE TABLE IF NOT EXISTS refresh_tokens (
				id BIGSERIAL PRIMARY KEY,
				user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				token_hash TEXT NOT NULL UNIQUE,
				user_agent TEXT NOT NULL DEFAULT '',
				ip_address TEXT NOT NULL DEFAULT '',
				is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
				expires_at TIMESTAMPTZ NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
			`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id)`,
			`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens (expires_at)`,
		},
	}
}

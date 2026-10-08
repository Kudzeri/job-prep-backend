package migrations

func AssistantRequests() SchemaFile {
	return SchemaFile{Name: "assistant_requests", Statements: []string{
		`CREATE TABLE IF NOT EXISTS assistant_requests (
			id UUID PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed')),
			response_message TEXT NOT NULL DEFAULT '',
		test_id BIGINT REFERENCES tests(id) ON DELETE SET NULL,
			error_message TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS assistant_requests_user_created_idx ON assistant_requests(user_id, created_at DESC)`,
	}}
}

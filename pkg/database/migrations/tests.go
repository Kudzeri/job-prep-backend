package migrations

func Tests() SchemaFile {
	return SchemaFile{Name: "tests", Statements: []string{
		`CREATE TABLE IF NOT EXISTS tests (
			id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			owner_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`ALTER TABLE tests ADD COLUMN IF NOT EXISTS owner_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL`,
		`CREATE INDEX IF NOT EXISTS tests_owner_user_id_idx ON tests(owner_user_id)`,
		`CREATE TABLE IF NOT EXISTS test_questions (
			id BIGSERIAL PRIMARY KEY,
			test_id BIGINT NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
			text TEXT NOT NULL,
			options JSONB,
			answer JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS test_questions_test_id_idx ON test_questions(test_id)`,
	}}
}

package migrations

func EmailOTPCodes() SchemaFile {
	return SchemaFile{
		Name: "email_otp_codes",
		Statements: []string{
			`CREATE TABLE IF NOT EXISTS email_otp_codes (
				id BIGSERIAL PRIMARY KEY,
				email TEXT NOT NULL,
				code TEXT NOT NULL,
				is_used BOOLEAN NOT NULL DEFAULT FALSE,
				expires_at TIMESTAMPTZ NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
			`CREATE INDEX IF NOT EXISTS idx_email_otp_codes_email_code ON email_otp_codes (email, code)`,
			`CREATE INDEX IF NOT EXISTS idx_email_otp_codes_expires_at ON email_otp_codes (expires_at)`,
		},
	}
}

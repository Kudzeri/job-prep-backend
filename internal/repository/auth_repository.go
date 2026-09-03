package repository

import (
	"context"
	"time"

	"github.com/Kudzeri/job-prep-backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{db: db}
}

// --- Email OTP ---

func (r *AuthRepository) SaveEmailOTP(ctx context.Context, email, code string, expiresAt time.Time) error {
	query := `
		INSERT INTO email_otp_codes (email, code, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, email, code, expiresAt)
	return err
}

func (r *AuthRepository) GetValidEmailOTP(ctx context.Context, email, code string) (*domain.EmailOTP, error) {
	query := `
		SELECT id, email, code, is_used, expires_at
		FROM email_otp_codes
		WHERE email = $1 AND code = $2 AND is_used = FALSE AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`
	var otp domain.EmailOTP
	err := r.db.QueryRow(ctx, query, email, code).Scan(
		&otp.ID, &otp.Email, &otp.Code, &otp.IsUsed, &otp.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *AuthRepository) MarkOTPAsUsed(ctx context.Context, otpID int64) error {
	query := `UPDATE email_otp_codes SET is_used = TRUE WHERE id = $1`
	_, err := r.db.Exec(ctx, query, otpID)
	return err
}

// --- User ---

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, email_verified, telegram_id, telegram_username, first_name, is_onboarded, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.TelegramID,
		&user.TelegramUsername, &user.FirstName, &user.IsOnboarded,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) CreateUserWithEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		INSERT INTO users (email, email_verified)
		VALUES ($1, TRUE)
		RETURNING id, email, email_verified, telegram_id, telegram_username, first_name, is_onboarded, role, is_active, created_at, updated_at
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.TelegramID,
		&user.TelegramUsername, &user.FirstName, &user.IsOnboarded,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// --- Refresh Token ---

func (r *AuthRepository) SaveRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, token.UserID, token.TokenHash, token.UserAgent, token.IPAddress, token.ExpiresAt)
	return err
}

// --- Telegram Auth Sessions ---

func (r *AuthRepository) CreateTelegramSession(ctx context.Context, authToken string, expiresAt time.Time) error {
	query := `
		INSERT INTO telegram_auth_sessions (auth_token, status, expires_at)
		VALUES ($1, 'pending', $2)
	`
	_, err := r.db.Exec(ctx, query, authToken, expiresAt)
	return err
}

func (r *AuthRepository) GetTelegramSession(ctx context.Context, authToken string) (*domain.TelegramAuthSession, error) {
	query := `
		SELECT id, auth_token, telegram_id, status, expires_at, created_at
		FROM telegram_auth_sessions
		WHERE auth_token = $1 AND expires_at > NOW()
	`
	var session domain.TelegramAuthSession
	err := r.db.QueryRow(ctx, query, authToken).Scan(
		&session.ID, &session.AuthToken, &session.TelegramID, &session.Status, &session.ExpiresAt, &session.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *AuthRepository) ConfirmTelegramSession(ctx context.Context, authToken string, telegramID int64) error {
	query := `
		UPDATE telegram_auth_sessions
		SET telegram_id = $1, status = 'approved'
		WHERE auth_token = $2 AND status = 'pending' AND expires_at > NOW()
	`
	_, err := r.db.Exec(ctx, query, telegramID, authToken)
	return err
}

// --- Telegram User ---

func (r *AuthRepository) GetUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	query := `
		SELECT id, email, email_verified, telegram_id, telegram_username, first_name, is_onboarded, role, is_active, created_at, updated_at
		FROM users
		WHERE telegram_id = $1
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, telegramID).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.TelegramID,
		&user.TelegramUsername, &user.FirstName, &user.IsOnboarded,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) CreateUserWithTelegram(ctx context.Context, telegramID int64, username string) (*domain.User, error) {
	query := `
		INSERT INTO users (telegram_id, telegram_username)
		VALUES ($1, $2)
		RETURNING id, email, email_verified, telegram_id, telegram_username, first_name, is_onboarded, role, is_active, created_at, updated_at
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, telegramID, username).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.TelegramID,
		&user.TelegramUsername, &user.FirstName, &user.IsOnboarded,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
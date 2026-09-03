package repository

import (
	"context"

	"github.com/Kudzeri/job-prep-backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID получает текущий профиль пользователя
func (r *UserRepository) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	query := `
		SELECT id, email, email_verified, telegram_id, telegram_username, first_name, is_onboarded, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = TRUE
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.TelegramID,
		&user.TelegramUsername, &user.FirstName, &user.IsOnboarded,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CompleteProfile завершает Onboarding: устанавливает имя и флаг is_onboarded = TRUE
func (r *UserRepository) CompleteProfile(ctx context.Context, userID int64, firstName string) (*domain.User, error) {
	query := `
		UPDATE users
		SET first_name = $1, is_onboarded = TRUE, updated_at = NOW()
		WHERE id = $2
		RETURNING id, email, email_verified, telegram_id, telegram_username, first_name, is_onboarded, role, is_active, created_at, updated_at
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, firstName, userID).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.TelegramID,
		&user.TelegramUsername, &user.FirstName, &user.IsOnboarded,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
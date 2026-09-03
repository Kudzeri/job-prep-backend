package domain

import "time"

type User struct {
	ID               int64     `json:"id"`
	Email            *string   `json:"email"`
	EmailVerified    bool      `json:"email_verified"`
	TelegramID       *int64    `json:"telegram_id"`
	TelegramUsername *string   `json:"telegram_username"`
	FirstName        *string   `json:"first_name"`
	IsOnboarded      bool      `json:"is_onboarded"`
	Role             string    `json:"role"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type RefreshToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	TokenHash string    `json:"-"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	IsRevoked bool      `json:"is_revoked"`
	ExpiresAt time.Time `json:"expires_at"`
}

type EmailOTP struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	IsUsed    bool      `json:"is_used"`
	ExpiresAt time.Time `json:"expires_at"`
}

type TelegramAuthSession struct {
	ID         int64     `json:"id"`
	AuthToken  string    `json:"auth_token"`
	TelegramID *int64    `json:"telegram_id"`
	Status     string    `json:"status"` // "pending", "approved", "expired"
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}
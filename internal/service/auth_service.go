package service

import (
	"context"
	"crypto/rand"

	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/Kudzeri/job-prep-backend/internal/domain"
	"github.com/Kudzeri/job-prep-backend/internal/repository"
	"github.com/Kudzeri/job-prep-backend/pkg/email"
	"github.com/Kudzeri/job-prep-backend/pkg/jwt"
)

type AuthService struct {
	repo        *repository.AuthRepository
	emailSender *email.Sender
	jwtSecret   string
}

func NewAuthService(repo *repository.AuthRepository, emailSender *email.Sender, jwtSecret string) *AuthService {
	return &AuthService{
		repo:        repo,
		emailSender: emailSender,
		jwtSecret:   jwtSecret,
	}
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// SendOTP генерирует 6-значный код и отправляет на Email
func (s *AuthService) SendOTP(ctx context.Context, toEmail string) error {
	// 1. Генерируем 6-значный случайный код
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", n.Int64()+100000)

	// 2. Сохраняем в БД на 5 минут
	expiresAt := time.Now().Add(5 * time.Minute)
	if err := s.repo.SaveEmailOTP(ctx, toEmail, code, expiresAt); err != nil {
		return err
	}

	// 3. Реальная отправка письма через SMTP
	if s.emailSender != nil {
		if err := s.emailSender.SendOTP(toEmail, code); err != nil {
			return fmt.Errorf("ошибка отправки email: %w", err)
		}
	}

	return nil
}

// VerifyOTP проверяет код, регистрирует/находит пользователя и выдаёт JWT токены
func (s *AuthService) VerifyOTP(ctx context.Context, email, code, userAgent, ip string) (*AuthTokens, error) {
	// 1. Проверяем валидность OTP
	otp, err := s.repo.GetValidEmailOTP(ctx, email, code)
	if err != nil {
		return nil, errors.New("неверный или истекший OTP код")
	}

	// 2. Помечаем OTP как использованный
	if err := s.repo.MarkOTPAsUsed(ctx, otp.ID); err != nil {
		return nil, err
	}

	// 3. Находим или создаем пользователя
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		user, err = s.repo.CreateUserWithEmail(ctx, email)
		if err != nil {
			return nil, err
		}
	}

	// 4. Генерируем Access и Refresh токены
	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshTokenStr, err := jwt.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	// 5. Хэшируем Refresh-токен и сохраняем в БД
	hash := sha256.Sum256([]byte(refreshTokenStr))
	tokenHash := hex.EncodeToString(hash[:])

	refreshTokenObj := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		UserAgent: userAgent,
		IPAddress: ip,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}

	if err := s.repo.SaveRefreshToken(ctx, refreshTokenObj); err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
	}, nil
}

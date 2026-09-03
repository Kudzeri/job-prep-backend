package service

import (
	"context"
	"errors"

	"github.com/Kudzeri/job-prep-backend/internal/domain"
	"github.com/Kudzeri/job-prep-backend/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (*domain.User, error) {
	return s.repo.GetByID(ctx, userID)
}

type UpdateProfileInput struct {
	FirstName *string `json:"first_name"`
}

// CompleteProfile используется только при первом входе (Onboarding)
func (s *UserService) CompleteProfile(ctx context.Context, userID int64, firstName string) (*domain.User, error) {
	if firstName == "" {
		return nil, errors.New("имя не может быть пустым")
	}

	return s.repo.CompleteProfile(ctx, userID, firstName)
}

// UpdateProfile используется для последующего редактирования профиля
func (s *UserService) UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (*domain.User, error) {
	if input.FirstName != nil && *input.FirstName == "" {
		return nil, errors.New("имя не может быть пустым")
	}

	return s.repo.UpdateProfile(ctx, userID, input.FirstName)
}
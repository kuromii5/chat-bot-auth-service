package user

//go:generate mockery

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	UpdatePreferences(ctx context.Context, userID uuid.UUID, emailEnabled bool) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type Service struct {
	repo UserRepo
}

func NewService(repo UserRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpdatePreferences(
	ctx context.Context,
	userID uuid.UUID,
	emailEnabled bool,
) error {
	if err := s.repo.UpdatePreferences(ctx, userID, emailEnabled); err != nil {
		return fmt.Errorf("failed to update preferences: %w", err)
	}
	return nil
}

func (s *Service) GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}
	return user, nil
}

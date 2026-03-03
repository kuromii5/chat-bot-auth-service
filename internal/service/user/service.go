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
}

type Service struct {
	repo UserRepo
}

func NewService(repo UserRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpdatePreferences(ctx context.Context, userID uuid.UUID, emailEnabled bool) error {
	if err := s.repo.UpdatePreferences(ctx, userID, emailEnabled); err != nil {
		return fmt.Errorf("failed to update preferences: %w", err)
	}
	return nil
}

package user

//go:generate mockery

import (
	"context"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
}

type Service struct {
	repo UserRepo
}

func NewService(repo UserRepo) *Service {
	return &Service{repo: repo}
}

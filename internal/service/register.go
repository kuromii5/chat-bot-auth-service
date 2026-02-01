package service

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

type RegisterRequest struct {
	Email    string
	Password string
	Username string
	Role     domain.Role
}

type RegisterResponse struct {
	UserID   uuid.UUID
	Email    string
	Username string
	Role     domain.Role
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.Create(ctx, &domain.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Username:     req.Username,
		Role:         req.Role,
	})
	if err != nil {
		return nil, err
	}

	return &RegisterResponse{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
	}, nil
}

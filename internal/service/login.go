package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
	"github.com/kuromii5/chat-bot-auth-service/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

type LoginResponse struct {
	AccessToken  string
	RefreshToken string
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, err := s.jwtManager.GenerateAccess(user.ID)
	if err != nil {
		return nil, err
	}

	refreshTokenStr := uuid.New().String()
	hashedToken := jwt.SHA256(refreshTokenStr)

	refresh := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashedToken,
		UserAgent: &req.UserAgent,
		IPAddress: &req.IPAddress,
		ExpiresAt: time.Now().Add(s.jwtManager.RefreshTokenExpiry()),
	}

	if err := s.tokenRepo.Create(ctx, refresh); err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	hashedToken := jwt.SHA256(refreshToken)
	return s.tokenRepo.Revoke(ctx, hashedToken)
}

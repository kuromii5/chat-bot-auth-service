package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
	"github.com/kuromii5/chat-bot-auth-service/pkg/jwt"
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
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, err := s.jwtManager.GenerateAccess(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
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

	if err := s.tokenRepo.CreateToken(ctx, refresh); err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	hashedToken := jwt.SHA256(refreshToken)
	if err := s.tokenRepo.RevokeToken(ctx, hashedToken); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}

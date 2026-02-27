package session

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

type RefreshRequest struct {
	OldRefreshTokenRaw string
	UserAgent          string
	IPAddress          string
}

type RefreshResponse struct {
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

	if err := s.tokenRepo.CreateToken(ctx, &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashedToken,
		UserAgent: &req.UserAgent,
		IPAddress: &req.IPAddress,
		ExpiresAt: time.Now().Add(s.jwtManager.RefreshTokenExpiry()),
	}); err != nil {
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

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*RefreshResponse, error) {
	oldHash := jwt.SHA256(req.OldRefreshTokenRaw)

	tokenDoc, err := s.tokenRepo.GetToken(ctx, oldHash)
	if err != nil || tokenDoc == nil {
		return nil, domain.ErrTokenNotFound
	}

	if tokenDoc.RevokedAt != nil {
		return nil, domain.ErrTokenRevoked
	}

	if time.Now().After(tokenDoc.ExpiresAt) {
		return nil, domain.ErrTokenExpired
	}

	newAccessToken, err := s.jwtManager.GenerateAccess(tokenDoc.UserID, tokenDoc.Role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshTokenRaw := uuid.New().String()
	newHash := jwt.SHA256(newRefreshTokenRaw)

	if err := s.tokenRepo.RevokeToken(ctx, oldHash); err != nil {
		return nil, fmt.Errorf("revoke old token: %w", err)
	}

	if err := s.tokenRepo.CreateToken(ctx, &domain.RefreshToken{
		UserID:    tokenDoc.UserID,
		TokenHash: newHash,
		UserAgent: &req.UserAgent,
		IPAddress: &req.IPAddress,
		ExpiresAt: time.Now().Add(s.jwtManager.RefreshTokenExpiry()),
	}); err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	return &RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenRaw,
	}, nil
}

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
	"github.com/kuromii5/chat-bot-auth-service/pkg/jwt"
)

type RefreshRequest struct {
	OldRefreshTokenRaw string
	UserAgent          string
	IPAddress          string
}

type RefreshResponse struct {
	AccessToken  string
	RefreshToken string
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*RefreshResponse, error) {
	oldHash := jwt.SHA256(req.OldRefreshTokenRaw)

	tokenDoc, err := s.tokenRepo.GetToken(ctx, oldHash)
	if err != nil || tokenDoc == nil {
		return nil, domain.ErrTokenNotFound
	}

	if tokenDoc.IsRevoked {
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
		return nil, fmt.Errorf("revoke token: %w", err)
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

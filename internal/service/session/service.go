package session

//go:generate mockery

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

type UserRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type TokenRepo interface {
	CreateToken(ctx context.Context, token *domain.RefreshToken) error
	GetToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeToken(ctx context.Context, tokenHash string) error
}

type JWTManager interface {
	GenerateAccess(userID uuid.UUID, role domain.Role) (string, error)
	RefreshTokenExpiry() time.Duration
}

type Service struct {
	userRepo   UserRepo
	tokenRepo  TokenRepo
	jwtManager JWTManager
}

func NewService(userRepo UserRepo, tokenRepo TokenRepo, jwtManager JWTManager) *Service {
	return &Service{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
	}
}

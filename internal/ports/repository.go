package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
}

type RefreshTokenRepository interface {
	CreateToken(ctx context.Context, token *domain.RefreshToken) error
	GetToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeToken(ctx context.Context, tokenHash string) error
	RevokeAllTokens(ctx context.Context, userID uuid.UUID) error
}

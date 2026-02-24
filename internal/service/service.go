package service

import (
	"github.com/kuromii5/chat-bot-auth-service/internal/ports"
	"github.com/kuromii5/chat-bot-auth-service/pkg/jwt"
)

type Service struct {
	userRepo   ports.UserRepository
	tokenRepo  ports.RefreshTokenRepository
	jwtManager *jwt.JWTManager
}

func NewService(
	userRepo ports.UserRepository,
	tokenRepo ports.RefreshTokenRepository,
	jwtManager *jwt.JWTManager,
) *Service {
	return &Service{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
	}
}

package service

import (
	"github.com/kuromii5/chat-bot-auth-service/internal/ports"
)

type Service struct {
	userRepo ports.UserRepository
}

func NewService(userRepo ports.UserRepository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

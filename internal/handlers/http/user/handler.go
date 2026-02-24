package user

import (
	"context"

	"github.com/kuromii5/chat-bot-auth-service/internal/service/user"
)

type Service interface {
	Register(ctx context.Context, req user.RegisterRequest) (*user.RegisterResponse, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

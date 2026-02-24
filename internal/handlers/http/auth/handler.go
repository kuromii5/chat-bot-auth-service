package auth

import (
	"context"

	"github.com/kuromii5/chat-bot-auth-service/internal/service/session"
)

type Service interface {
	Login(ctx context.Context, req session.LoginRequest) (*session.LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, req session.RefreshRequest) (*session.RefreshResponse, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

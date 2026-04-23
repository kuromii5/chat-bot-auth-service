package user

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
	authv1 "github.com/kuromii5/chat-bot-shared/proto/auth/v1"
)

type UserService interface {
	GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.User, error)
}

type Handler struct {
	authv1.UnimplementedUserServiceServer
	svc UserService
}

func NewHandler(svc UserService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetPreferences(ctx context.Context, req *authv1.GetPreferencesRequest) (*authv1.GetPreferencesResponse, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	user, err := h.svc.GetPreferences(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &authv1.GetPreferencesResponse{
		Email:                     user.Email,
		EmailNotificationsEnabled: user.EmailNotificationsEnabled,
	}, nil
}

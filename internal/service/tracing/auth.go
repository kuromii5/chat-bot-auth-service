package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/kuromii5/chat-bot-auth-service/internal/service/session"
)

type authInner interface {
	Login(ctx context.Context, req session.LoginRequest) (*session.LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, req session.RefreshRequest) (*session.RefreshResponse, error)
}

// AuthService wraps the session service and adds OTel spans around each method.
// It satisfies handlers/http/auth.Service via duck typing.
type AuthService struct {
	inner authInner
}

func NewAuthService(inner authInner) *AuthService {
	return &AuthService{inner: inner}
}

func (s *AuthService) Login(ctx context.Context, req session.LoginRequest) (*session.LoginResponse, error) {
	ctx, span := otel.Tracer("service/session").Start(ctx, "session.Login")
	defer span.End()
	span.SetAttributes(attribute.String("user.email", req.Email))

	result, err := s.inner.Login(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	ctx, span := otel.Tracer("service/session").Start(ctx, "session.Logout")
	defer span.End()

	err := s.inner.Logout(ctx, refreshToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (s *AuthService) Refresh(ctx context.Context, req session.RefreshRequest) (*session.RefreshResponse, error) {
	ctx, span := otel.Tracer("service/session").Start(ctx, "session.Refresh")
	defer span.End()

	result, err := s.inner.Refresh(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

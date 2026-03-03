package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	userservice "github.com/kuromii5/chat-bot-auth-service/internal/service/user"
)

type userInner interface {
	Register(
		ctx context.Context,
		req userservice.RegisterRequest,
	) (*userservice.RegisterResponse, error)
}

// UserService wraps the user service and adds an OTel span around Register.
// It satisfies handlers/http/user.Service via duck typing.
type UserService struct {
	inner userInner
}

func NewUserService(inner userInner) *UserService {
	return &UserService{inner: inner}
}

func (s *UserService) Register(
	ctx context.Context,
	req userservice.RegisterRequest,
) (*userservice.RegisterResponse, error) {
	ctx, span := otel.Tracer("service/user").Start(ctx, "user.Register")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.email", req.Email),
		attribute.String("user.role", string(req.Role)),
	)

	result, err := s.inner.Register(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

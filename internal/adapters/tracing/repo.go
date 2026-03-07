package tracing

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

// postgresRepo is the union of all repo interfaces defined in the service layer.
// TracingRepo satisfies them all via duck typing — no service package imports needed.
type postgresRepo interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	UpdatePreferences(ctx context.Context, userID uuid.UUID, emailEnabled bool) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateToken(ctx context.Context, token *domain.RefreshToken) error
	GetToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeToken(ctx context.Context, tokenHash string) error
}

const dbTracer = "postgres"

// Repo wraps any postgresRepo and adds an OTel span to every DB call.
// The business logic (service layer, postgres adapter) stays OTel-free.
type Repo struct {
	inner postgresRepo
}

func NewRepo(inner postgresRepo) *Repo {
	return &Repo{inner: inner}
}

func (r *Repo) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.CreateUser")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "auth.users"),
		attribute.String("user.email", user.Email),
		attribute.String("user.role", string(user.Role)),
	)

	result, err := r.inner.CreateUser(ctx, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (r *Repo) UpdatePreferences(ctx context.Context, userID uuid.UUID, emailEnabled bool) error {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.UpdatePreferences")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.table", "auth.users"),
		attribute.String("user.id", userID.String()),
	)

	err := r.inner.UpdatePreferences(ctx, userID, emailEnabled)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (r *Repo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.GetUserByEmail")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "auth.users"),
	)

	result, err := r.inner.GetUserByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (r *Repo) CreateToken(ctx context.Context, token *domain.RefreshToken) error {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.CreateToken")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "auth.refresh_tokens"),
		attribute.String("user.id", token.UserID.String()),
	)

	err := r.inner.CreateToken(ctx, token)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (r *Repo) GetToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.GetToken")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "auth.refresh_tokens"),
	)

	result, err := r.inner.GetToken(ctx, tokenHash)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (r *Repo) RevokeToken(ctx context.Context, tokenHash string) error {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.RevokeToken")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.table", "auth.refresh_tokens"),
	)

	err := r.inner.RevokeToken(ctx, tokenHash)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

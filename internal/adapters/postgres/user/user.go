package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kuromii5/chat-bot-auth-service/internal/adapters/postgres"
	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

func (r *Repository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	var createdUser domain.User
	err := r.db.GetContext(ctx, &createdUser, createUserQuery,
		user.Email,
		user.PasswordHash,
		user.Username,
		user.Role,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == postgres.UniqueViolationErrorCode {
				return nil, fmt.Errorf("failed to create user: %w detail: %s", domain.ErrUserAlreadyExists, pgErr.Detail)
			}
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &createdUser, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, findByEmailQuery, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, findByUsernameQuery, username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

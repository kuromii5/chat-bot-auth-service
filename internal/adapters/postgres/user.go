package postgres

import (
	"context"
	"fmt"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

func (r *postgres) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // no-op if tx already committed
	}()

	if err := tx.GetContext(
		ctx,
		user,
		createAuthUserQuery,
		user.Email,
		user.PasswordHash,
		user.Role,
	); err != nil {
		return nil, r.handleError(err, "email")
	}

	if _, err := tx.ExecContext(ctx, createProfileQuery, user.ID, user.Username); err != nil {
		return nil, r.handleError(err, "username")
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user, nil
}

func (r *postgres) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.DB.GetContext(ctx, &user, getUserByEmailQuery, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

func (r *postgres) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.DB.GetContext(ctx, &user, getUserByUsernameQuery, username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

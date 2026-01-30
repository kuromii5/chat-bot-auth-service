package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

func (r *Repository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var authData struct {
		ID           uuid.UUID `db:"id"`
		CreatedAt    time.Time `db:"created_at"`
		TokenVersion int       `db:"token_version"`
	}

	if err := tx.QueryRowxContext(ctx, createAuthUserQuery, user.Email, user.PasswordHash).StructScan(&authData); err != nil {
		return nil, r.handleError(err, "email")
	}

	if _, err := tx.ExecContext(ctx, createProfileQuery, authData.ID, user.Username, user.Role); err != nil {
		return nil, r.handleError(err, "username")
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	user.ID = authData.ID
	user.CreatedAt = authData.CreatedAt
	user.TokenVersion = authData.TokenVersion
	return user, nil
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

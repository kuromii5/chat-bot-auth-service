package token

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

func (r *Repository) Create(ctx context.Context, t *domain.RefreshToken) error {
	_, err := r.db.ExecContext(ctx, createRefreshTokenQuery, t.UserID, t.TokenHash, t.UserAgent, t.IPAddress, t.ExpiresAt)
	return err
}

func (r *Repository) Revoke(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, revokeRefreshTokenQuery, tokenHash)
	return err
}

func (r *Repository) Get(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var t domain.RefreshToken

	err := r.db.GetContext(ctx, &t, getRefreshTokenQuery, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}

	return &t, nil
}

func (r *Repository) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, revokeAllTokensQuery, userID)
	return err
}

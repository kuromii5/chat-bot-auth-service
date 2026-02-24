package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

func (r *postgres) CreateToken(ctx context.Context, t *domain.RefreshToken) error {
	_, err := r.DB.ExecContext(
		ctx,
		createRefreshTokenQuery,
		t.UserID,
		t.TokenHash,
		t.UserAgent,
		t.IPAddress,
		t.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create token: %w", err)
	}
	return nil
}

func (r *postgres) RevokeToken(ctx context.Context, tokenHash string) error {
	_, err := r.DB.ExecContext(ctx, revokeRefreshTokenQuery, tokenHash)
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}

func (r *postgres) GetToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var t domain.RefreshToken

	err := r.DB.GetContext(ctx, &t, getRefreshTokenQuery, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("get token: %w", err)
	}

	return &t, nil
}

func (r *postgres) RevokeAllTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := r.DB.ExecContext(ctx, revokeAllTokensQuery, userID)
	if err != nil {
		return fmt.Errorf("revoke all tokens: %w", err)
	}
	return nil
}

package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

func (r *Postgres) CreateToken(ctx context.Context, t *domain.RefreshToken) error {
	_, err := r.DB.ExecContext(ctx, createRefreshTokenQuery, t.UserID, t.TokenHash, t.UserAgent, t.IPAddress, t.ExpiresAt)
	return err
}

func (r *Postgres) RevokeToken(ctx context.Context, tokenHash string) error {
	_, err := r.DB.ExecContext(ctx, revokeRefreshTokenQuery, tokenHash)
	return err
}

func (r *Postgres) GetToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var t domain.RefreshToken

	err := r.DB.GetContext(ctx, &t, getRefreshTokenQuery, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}

	return &t, nil
}

func (r *Postgres) RevokeAllTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := r.DB.ExecContext(ctx, revokeAllTokensQuery, userID)
	return err
}

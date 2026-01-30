package user

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

var UniqueViolationErrorCode = "23505"

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) handleError(err error, field string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == UniqueViolationErrorCode {
			return fmt.Errorf("failed to create user: %w (field: %s)", domain.ErrUserAlreadyExists, field)
		}
	}
	return fmt.Errorf("database error: %w", err)
}

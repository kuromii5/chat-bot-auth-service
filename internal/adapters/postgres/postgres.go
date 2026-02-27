package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/kuromii5/chat-bot-auth-service/config"
	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

const UniqueViolationErrorCode = "23505"

type postgres struct {
	DB *sqlx.DB
}

func New(cfg *config.DatabaseConfig) (*postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, "pgx", DSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &postgres{DB: db}, nil
}

func DSN(c *config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func (r *postgres) handleError(err error, field string) error {
	if pgerr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgerr.Code == UniqueViolationErrorCode {
			return fmt.Errorf(
				"failed to create user: %w (field: %s)",
				domain.ErrUserAlreadyExists,
				field,
			)
		}
	}
	return fmt.Errorf("database error: %w", err)
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	AI    Role = "AI"
	Human Role = "Human"
)

type User struct {
	ID                       uuid.UUID  `db:"id"`
	Email                    string     `db:"email"`
	PasswordHash             string     `db:"password_hash"`
	Username                 string     `db:"username"`
	Role                     Role       `db:"role"`
	IsVerified               bool       `db:"is_verified"`
	TokenVersion             int        `db:"token_version"`
	EmailNotificationsEnabled bool      `db:"email_notifications_enabled"`
	CreatedAt                time.Time  `db:"created_at"`
	UpdatedAt                time.Time  `db:"updated_at"`
	DeletedAt                *time.Time `db:"deleted_at"`
}

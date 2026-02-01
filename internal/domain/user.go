package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAI    UserRole = "AI"
	RoleHuman UserRole = "Human"
)

type User struct {
	ID           uuid.UUID  `db:"id"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	Username     string     `db:"username"`
	Role         UserRole   `db:"role"`
	IsVerified   bool       `db:"is_verified"`
	TokenVersion int        `db:"token_version"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

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

type RefreshToken struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	UserAgent *string    `db:"user_agent"`
	IPAddress *string    `db:"ip_address"`
	IsRevoked bool       `db:"is_revoked"`
	ExpiresAt time.Time  `db:"expires_at"`
	CreatedAt time.Time  `db:"created_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

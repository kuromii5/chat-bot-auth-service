package domain

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenRevoked       = errors.New("token revoked")

	ErrAuthorizationHeaderRequired = errors.New("authorization header required")
	ErrInvalidAuthorizationFormat  = errors.New("invalid authorization format")
	ErrInvalidOrExpiredToken       = errors.New("invalid or expired token")
)

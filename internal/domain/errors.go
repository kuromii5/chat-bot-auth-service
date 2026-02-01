package domain

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidParameters  = errors.New("invalid parameters")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenRevoked       = errors.New("token revoked")
)

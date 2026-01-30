package domain

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidParameters = errors.New("invalid parameters")
)

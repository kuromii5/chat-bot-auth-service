package errors

import (
	"net/http"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
	"github.com/kuromii5/chat-bot-shared/wrapper"
)

var Registry = map[error]wrapper.ErrorResponse{
	domain.ErrUserAlreadyExists:  {Status: http.StatusConflict, Message: "User with this email or username already exists"},
	domain.ErrInvalidCredentials: {Status: http.StatusUnauthorized, Message: "Invalid credentials"},
	domain.ErrTokenNotFound:      {Status: http.StatusNotFound, Message: "Token not found"},
	domain.ErrTokenExpired:       {Status: http.StatusUnauthorized, Message: "Token expired"},
	domain.ErrTokenRevoked:       {Status: http.StatusUnauthorized, Message: "Token revoked"},
}

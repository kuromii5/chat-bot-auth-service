package http

import (
	"encoding/json"
	"net/http"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
	"github.com/kuromii5/chat-bot-auth-service/internal/service"
	"github.com/kuromii5/chat-bot-auth-service/pkg/validator"
	"github.com/kuromii5/chat-bot-auth-service/pkg/wrapper"
)

type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Username string `json:"username" validate:"required,min=3,max=32"`
	Role     string `json:"role" validate:"required,oneof=AI Human"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}
	if err := validator.Validate(req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	role := domain.UserRole(req.Role)
	registerReq := service.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
		Role:     role,
	}

	resp, err := h.service.Register(r.Context(), registerReq)
	if err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.JSON(w, http.StatusCreated, resp)
}

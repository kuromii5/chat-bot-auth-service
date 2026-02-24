package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kuromii5/chat-bot-auth-service/internal/service"
	"github.com/kuromii5/chat-bot-auth-service/pkg/wrapper"
)

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	userAgent := r.Header.Get("User-Agent")

	loginResp, err := h.service.Login(r.Context(), service.LoginRequest{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: userAgent,
		IPAddress: ip,
	})
	if err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.JSON(w, http.StatusOK, loginResponse{
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.NoContent(w)
}

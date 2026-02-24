package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kuromii5/chat-bot-auth-service/internal/service/session"
	"github.com/kuromii5/chat-bot-auth-service/pkg/validator"
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

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}
	if err := validator.Validate(req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}

	resp, err := h.svc.Login(r.Context(), session.LoginRequest{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: r.Header.Get("User-Agent"),
		IPAddress: ip,
	})
	if err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.JSON(w, http.StatusOK, loginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}
	if err := validator.Validate(req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.NoContent(w)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}
	if err := validator.Validate(req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}

	resp, err := h.svc.Refresh(r.Context(), session.RefreshRequest{
		OldRefreshTokenRaw: req.RefreshToken,
		UserAgent:          r.Header.Get("User-Agent"),
		IPAddress:          ip,
	})
	if err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.JSON(w, http.StatusOK, refreshResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	})
}

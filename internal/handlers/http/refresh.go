package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kuromii5/chat-bot-auth-service/internal/service"
	"github.com/kuromii5/chat-bot-auth-service/pkg/wrapper"
)

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	userAgent := r.Header.Get("User-Agent")

	resp, err := h.service.Refresh(r.Context(), service.RefreshRequest{
		OldRefreshTokenRaw: req.RefreshToken,
		UserAgent:          userAgent,
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

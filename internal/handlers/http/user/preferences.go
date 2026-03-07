package user

import (
	"encoding/json"
	"net/http"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
	authmw "github.com/kuromii5/chat-bot-auth-service/internal/handlers/http/middleware"
	"github.com/kuromii5/chat-bot-shared/wrapper"
)

type preferencesRequest struct {
	EmailNotificationsEnabled bool `json:"email_notifications_enabled"`
}

func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := authmw.UserIDFromCtx(r.Context())
	if !ok {
		wrapper.WrapError(w, r, domain.ErrInvalidOrExpiredToken)
		return
	}

	var req preferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	if err := h.svc.UpdatePreferences(
		r.Context(),
		userID,
		req.EmailNotificationsEnabled,
	); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.NoContent(w)
}

package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
	"github.com/kuromii5/chat-bot-auth-service/internal/handlers/http/auth/mocks"
	"github.com/kuromii5/chat-bot-auth-service/internal/service/session"
	"github.com/kuromii5/chat-bot-auth-service/pkg/validator"
)

func TestMain(m *testing.M) {
	validator.Init()
	os.Exit(m.Run())
}

// ─── Login ────────────────────────────────────────────────────────────────────

func TestLogin(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		headers    map[string]string
		setup      func(svc *mocks.MockService)
		wantStatus int
		wantKey    string
		wantVal    any
	}{
		{
			name: "success",
			body: `{"email":"alice@example.com","password":"secret123"}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Login(mock.Anything, mock.Anything).
					Return(&session.LoginResponse{
						AccessToken:  "access.token",
						RefreshToken: "refresh.token",
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantKey:    "access_token",
			wantVal:    "access.token",
		},
		{
			name:       "invalid json",
			body:       `{bad`,
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Invalid JSON body",
		},
		{
			name:       "validation: missing email",
			body:       `{"password":"secret123"}`,
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Validation failed",
		},
		{
			name: "service: invalid credentials",
			body: `{"email":"alice@example.com","password":"wrong"}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Login(mock.Anything, mock.Anything).
					Return(nil, domain.ErrInvalidCredentials)
			},
			wantStatus: http.StatusUnauthorized,
			wantKey:    "error",
			wantVal:    "Invalid credentials",
		},
		{
			name: "service: generic error",
			body: `{"email":"alice@example.com","password":"secret123"}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Login(mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("db unavailable"))
			},
			wantStatus: http.StatusInternalServerError,
			wantKey:    "error",
			wantVal:    "Internal server error",
		},
		{
			name:    "x-forwarded-for ip extraction",
			body:    `{"email":"alice@example.com","password":"secret123"}`,
			headers: map[string]string{"X-Forwarded-For": "203.0.113.50, 70.41.3.18"},
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Login(mock.Anything, mock.MatchedBy(func(req session.LoginRequest) bool {
					return req.IPAddress == "203.0.113.50"
				})).Return(&session.LoginResponse{
					AccessToken:  "access.token",
					RefreshToken: "refresh.token",
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantKey:    "access_token",
			wantVal:    "access.token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := NewHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if rec.Body.Len() > 0 {
				var body map[string]any
				err := json.NewDecoder(rec.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVal, body[tt.wantKey])
			}
		})
	}
}

// ─── Logout ───────────────────────────────────────────────────────────────────

func TestLogout(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(svc *mocks.MockService)
		wantStatus int
		wantKey    string
		wantVal    any
	}{
		{
			name: "success",
			body: `{"refresh_token":"some-token"}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Logout(mock.Anything, "some-token").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid json",
			body:       `{bad`,
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Invalid JSON body",
		},
		{
			name:       "validation: missing refresh_token",
			body:       `{}`,
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Validation failed",
		},
		{
			name: "service: error",
			body: `{"refresh_token":"some-token"}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Logout(mock.Anything, "some-token").
					Return(fmt.Errorf("db error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantKey:    "error",
			wantVal:    "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := NewHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Logout(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantKey != "" && rec.Body.Len() > 0 {
				var body map[string]any
				err := json.NewDecoder(rec.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVal, body[tt.wantKey])
			}
		})
	}
}

// ─── Refresh ──────────────────────────────────────────────────────────────────

func TestRefresh(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(svc *mocks.MockService)
		wantStatus int
		wantKey    string
		wantVal    any
	}{
		{
			name: "success",
			body: `{"refresh_token":"old-token"}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Refresh(mock.Anything, mock.Anything).
					Return(&session.RefreshResponse{
						AccessToken:  "new.access",
						RefreshToken: "new.refresh",
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantKey:    "access_token",
			wantVal:    "new.access",
		},
		{
			name:       "invalid json",
			body:       `{bad`,
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Invalid JSON body",
		},
		{
			name:       "validation: missing refresh_token",
			body:       `{}`,
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Validation failed",
		},
		{
			name: "service: token not found",
			body: `{"refresh_token":"old-token"}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Refresh(mock.Anything, mock.Anything).
					Return(nil, domain.ErrTokenNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantKey:    "error",
			wantVal:    "Token not found",
		},
		{
			name: "service: token expired",
			body: `{"refresh_token":"old-token"}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Refresh(mock.Anything, mock.Anything).
					Return(nil, domain.ErrTokenExpired)
			},
			wantStatus: http.StatusUnauthorized,
			wantKey:    "error",
			wantVal:    "Token expired",
		},
		{
			name: "service: token revoked",
			body: `{"refresh_token":"old-token"}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Refresh(mock.Anything, mock.Anything).
					Return(nil, domain.ErrTokenRevoked)
			},
			wantStatus: http.StatusUnauthorized,
			wantKey:    "error",
			wantVal:    "Token revoked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := NewHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Refresh(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if rec.Body.Len() > 0 {
				var body map[string]any
				err := json.NewDecoder(rec.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVal, body[tt.wantKey])
			}
		})
	}
}

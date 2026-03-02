package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
	userservice "github.com/kuromii5/chat-bot-auth-service/internal/service/user"
	"github.com/kuromii5/chat-bot-auth-service/internal/handlers/http/user/mocks"
	"github.com/kuromii5/chat-bot-auth-service/pkg/validator"
)

func TestMain(m *testing.M) {
	validator.Init()
	os.Exit(m.Run())
}

func TestRegister(t *testing.T) {
	userID := uuid.New()

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
			body: `{"email":"alice@example.com","password":"secret123","username":"alice","role":"Human"}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Register(mock.Anything, userservice.RegisterRequest{
					Email:    "alice@example.com",
					Password: "secret123",
					Username: "alice",
					Role:     domain.Human,
				}).Return(&userservice.RegisterResponse{
					UserID:   userID,
					Email:    "alice@example.com",
					Username: "alice",
					Role:     domain.Human,
				}, nil)
			},
			wantStatus: http.StatusCreated,
			wantKey:    "Email",
			wantVal:    "alice@example.com",
		},
		{
			name:       "invalid json body",
			body:       `{invalid`,
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Invalid JSON body",
		},
		{
			name:       "validation: missing required fields",
			body:       `{}`,
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Validation failed",
		},
		{
			name:       "validation: invalid email",
			body:       `{"email":"not-email","password":"secret123","username":"alice","role":"Human"}`,
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Validation failed",
		},
		{
			name:       "validation: short password",
			body:       `{"email":"alice@example.com","password":"short","username":"alice","role":"Human"}`,
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Validation failed",
		},
		{
			name:       "validation: invalid role",
			body:       `{"email":"alice@example.com","password":"secret123","username":"alice","role":"Admin"}`,
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Validation failed",
		},
		{
			name: "service: user already exists",
			body: `{"email":"alice@example.com","password":"secret123","username":"alice","role":"Human"}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Register(mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("failed: %w", domain.ErrUserAlreadyExists))
			},
			wantStatus: http.StatusConflict,
			wantKey:    "error",
			wantVal:    "User with this email or username already exists",
		},
		{
			name: "service: generic error",
			body: `{"email":"alice@example.com","password":"secret123","username":"alice","role":"Human"}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Register(mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("connection reset"))
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

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Register(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var body map[string]any
			err := json.NewDecoder(rec.Body).Decode(&body)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantVal, body[tt.wantKey])
		})
	}
}

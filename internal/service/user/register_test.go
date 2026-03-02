package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
	"github.com/kuromii5/chat-bot-auth-service/internal/service/user/mocks"
)

func TestRegister(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		req       RegisterRequest
		setup     func(repo *mocks.MockUserRepo)
		wantResp  *RegisterResponse
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "success",
			req: RegisterRequest{
				Email:    "alice@example.com",
				Password: "secret123",
				Username: "alice",
				Role:     domain.Human,
			},
			setup: func(repo *mocks.MockUserRepo) {
				repo.EXPECT().CreateUser(mock.Anything, mock.Anything).
					Return(&domain.User{
						ID:       userID,
						Email:    "alice@example.com",
						Username: "alice",
						Role:     domain.Human,
					}, nil)
			},
			wantResp: &RegisterResponse{
				UserID:   userID,
				Email:    "alice@example.com",
				Username: "alice",
				Role:     domain.Human,
			},
		},
		{
			name: "user already exists",
			req: RegisterRequest{
				Email:    "alice@example.com",
				Password: "secret123",
				Username: "alice",
				Role:     domain.Human,
			},
			setup: func(repo *mocks.MockUserRepo) {
				repo.EXPECT().CreateUser(mock.Anything, mock.Anything).
					Return(nil, domain.ErrUserAlreadyExists)
			},
			wantErr:   true,
			wantErrIs: domain.ErrUserAlreadyExists,
		},
		{
			name: "repo generic db failure",
			req: RegisterRequest{
				Email:    "bob@example.com",
				Password: "pass",
				Username: "bob",
				Role:     domain.AI,
			},
			setup: func(repo *mocks.MockUserRepo) {
				repo.EXPECT().CreateUser(mock.Anything, mock.Anything).
					Return(nil, errors.New("connection reset by peer"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockUserRepo(t)
			tt.setup(repo)

			svc := NewService(repo)
			resp, err := svc.Register(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp, resp)
			}
		})
	}
}

package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
	"github.com/kuromii5/chat-bot-auth-service/internal/service/session/mocks"
	jwtpkg "github.com/kuromii5/chat-bot-shared/jwt"
)

var (
	testUserID       = uuid.New()
	testPasswordHash = mustBcrypt("secret123")
)

func mustBcrypt(password string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	return string(h)
}

func newMocks(t *testing.T) (*mocks.MockUserRepo, *mocks.MockTokenRepo, *mocks.MockJWTManager) {
	return mocks.NewMockUserRepo(t), mocks.NewMockTokenRepo(t), mocks.NewMockJWTManager(t)
}

// ─── Login ────────────────────────────────────────────────────────────────────

func TestLogin(t *testing.T) {
	tests := []struct {
		name      string
		req       LoginRequest
		setup     func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager)
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "success",
			req: LoginRequest{
				Email:     "alice@example.com",
				Password:  "secret123",
				UserAgent: "Mozilla/5.0",
				IPAddress: "127.0.0.1",
			},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				ur.EXPECT().GetUserByEmail(mock.Anything, "alice@example.com").
					Return(&domain.User{
						ID:           testUserID,
						Email:        "alice@example.com",
						PasswordHash: testPasswordHash,
						Role:         domain.Human,
					}, nil)
				jm.EXPECT().GenerateAccess(testUserID, string(domain.Human)).Return("access.token.here", nil)
				jm.EXPECT().RefreshTokenExpiry().Return(24 * time.Hour)
				tr.EXPECT().CreateToken(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "user not found",
			req:  LoginRequest{Email: "ghost@example.com", Password: "secret123"},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				ur.EXPECT().GetUserByEmail(mock.Anything, "ghost@example.com").
					Return(nil, errors.New("no rows"))
			},
			wantErr:   true,
			wantErrIs: domain.ErrInvalidCredentials,
		},
		{
			name: "wrong password",
			req:  LoginRequest{Email: "alice@example.com", Password: "wrongpassword"},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				ur.EXPECT().GetUserByEmail(mock.Anything, "alice@example.com").
					Return(&domain.User{
						ID:           testUserID,
						PasswordHash: testPasswordHash,
						Role:         domain.Human,
					}, nil)
			},
			wantErr:   true,
			wantErrIs: domain.ErrInvalidCredentials,
		},
		{
			name: "generate access token fails",
			req:  LoginRequest{Email: "alice@example.com", Password: "secret123"},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				ur.EXPECT().GetUserByEmail(mock.Anything, "alice@example.com").
					Return(&domain.User{
						ID:           testUserID,
						PasswordHash: testPasswordHash,
						Role:         domain.Human,
					}, nil)
				jm.EXPECT().GenerateAccess(testUserID, string(domain.Human)).Return("", errors.New("signing error"))
			},
			wantErr: true,
		},
		{
			name: "create refresh token fails",
			req:  LoginRequest{Email: "alice@example.com", Password: "secret123"},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				ur.EXPECT().GetUserByEmail(mock.Anything, "alice@example.com").
					Return(&domain.User{
						ID:           testUserID,
						PasswordHash: testPasswordHash,
						Role:         domain.Human,
					}, nil)
				jm.EXPECT().GenerateAccess(testUserID, string(domain.Human)).Return("access.token.here", nil)
				jm.EXPECT().RefreshTokenExpiry().Return(24 * time.Hour)
				tr.EXPECT().CreateToken(mock.Anything, mock.Anything).Return(errors.New("db unavailable"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ur, tr, jm := newMocks(t)
			tt.setup(ur, tr, jm)
			svc := NewService(ur, tr, jm)

			resp, err := svc.Login(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			}
		})
	}
}

// ─── Logout ───────────────────────────────────────────────────────────────────

func TestLogout(t *testing.T) {
	const rawToken = "some-refresh-token-raw"
	tokenHash := jwtpkg.SHA256(rawToken)

	tests := []struct {
		name    string
		token   string
		setup   func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager)
		wantErr bool
	}{
		{
			name:  "success",
			token: rawToken,
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				tr.EXPECT().RevokeToken(mock.Anything, tokenHash).Return(nil)
			},
		},
		{
			name:  "repository unavailable",
			token: rawToken,
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				tr.EXPECT().RevokeToken(mock.Anything, tokenHash).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ur, tr, jm := newMocks(t)
			tt.setup(ur, tr, jm)
			svc := NewService(ur, tr, jm)

			err := svc.Logout(context.Background(), tt.token)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ─── Refresh ──────────────────────────────────────────────────────────────────

func TestRefresh(t *testing.T) {
	const rawToken = "old-refresh-token-raw"
	oldHash := jwtpkg.SHA256(rawToken)

	now := time.Now()
	pastTime := now.Add(-1 * time.Hour)

	validToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    testUserID,
		TokenHash: oldHash,
		Role:      domain.Human,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	revokedToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    testUserID,
		TokenHash: oldHash,
		Role:      domain.Human,
		ExpiresAt: now.Add(24 * time.Hour),
		RevokedAt: &pastTime,
	}
	expiredToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    testUserID,
		TokenHash: oldHash,
		Role:      domain.Human,
		ExpiresAt: pastTime,
	}

	tests := []struct {
		name      string
		req       RefreshRequest
		setup     func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager)
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "success",
			req:  RefreshRequest{OldRefreshTokenRaw: rawToken, UserAgent: "Mozilla/5.0", IPAddress: "127.0.0.1"},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				tr.EXPECT().GetToken(mock.Anything, oldHash).Return(validToken, nil)
				jm.EXPECT().GenerateAccess(testUserID, string(domain.Human)).Return("new.access.token", nil)
				tr.EXPECT().RevokeToken(mock.Anything, oldHash).Return(nil)
				jm.EXPECT().RefreshTokenExpiry().Return(24 * time.Hour)
				tr.EXPECT().CreateToken(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "token not found – repo error",
			req:  RefreshRequest{OldRefreshTokenRaw: rawToken},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				tr.EXPECT().GetToken(mock.Anything, oldHash).Return(nil, errors.New("not found"))
			},
			wantErr:   true,
			wantErrIs: domain.ErrTokenNotFound,
		},
		{
			name: "token not found – nil doc",
			req:  RefreshRequest{OldRefreshTokenRaw: rawToken},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				tr.EXPECT().GetToken(mock.Anything, oldHash).Return(nil, nil)
			},
			wantErr:   true,
			wantErrIs: domain.ErrTokenNotFound,
		},
		{
			name: "token revoked",
			req:  RefreshRequest{OldRefreshTokenRaw: rawToken},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				tr.EXPECT().GetToken(mock.Anything, oldHash).Return(revokedToken, nil)
			},
			wantErr:   true,
			wantErrIs: domain.ErrTokenRevoked,
		},
		{
			name: "token expired",
			req:  RefreshRequest{OldRefreshTokenRaw: rawToken},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				tr.EXPECT().GetToken(mock.Anything, oldHash).Return(expiredToken, nil)
			},
			wantErr:   true,
			wantErrIs: domain.ErrTokenExpired,
		},
		{
			name: "generate access token fails",
			req:  RefreshRequest{OldRefreshTokenRaw: rawToken},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				tr.EXPECT().GetToken(mock.Anything, oldHash).Return(validToken, nil)
				jm.EXPECT().GenerateAccess(testUserID, string(domain.Human)).Return("", errors.New("signing error"))
			},
			wantErr: true,
		},
		{
			name: "revoke old token fails",
			req:  RefreshRequest{OldRefreshTokenRaw: rawToken},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				tr.EXPECT().GetToken(mock.Anything, oldHash).Return(validToken, nil)
				jm.EXPECT().GenerateAccess(testUserID, string(domain.Human)).Return("new.access.token", nil)
				tr.EXPECT().RevokeToken(mock.Anything, oldHash).Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "create new refresh token fails",
			req:  RefreshRequest{OldRefreshTokenRaw: rawToken},
			setup: func(ur *mocks.MockUserRepo, tr *mocks.MockTokenRepo, jm *mocks.MockJWTManager) {
				tr.EXPECT().GetToken(mock.Anything, oldHash).Return(validToken, nil)
				jm.EXPECT().GenerateAccess(testUserID, string(domain.Human)).Return("new.access.token", nil)
				tr.EXPECT().RevokeToken(mock.Anything, oldHash).Return(nil)
				jm.EXPECT().RefreshTokenExpiry().Return(24 * time.Hour)
				tr.EXPECT().CreateToken(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ur, tr, jm := newMocks(t)
			tt.setup(ur, tr, jm)
			svc := NewService(ur, tr, jm)

			resp, err := svc.Refresh(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			}
		})
	}
}

//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

type tokenRepo interface {
	CreateToken(ctx context.Context, token *domain.RefreshToken) error
	GetToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeToken(ctx context.Context, tokenHash string) error
	RevokeAllTokens(ctx context.Context, userID uuid.UUID) error
}

// createTestUser is a helper that inserts a user and returns it.
func createTestUser(t *testing.T, email, username string) *domain.User {
	t.Helper()
	user, err := testRepo.CreateUser(context.Background(), &domain.User{
		Email:        email,
		PasswordHash: "hashed",
		Username:     username,
		Role:         domain.Human,
	})
	require.NoError(t, err)
	return user
}

func TestCreateToken_Success(t *testing.T) {
	truncateAll(t)
	user := createTestUser(t, "token@example.com", "tokenuser")

	ua := "Mozilla/5.0"
	ip := "127.0.0.1"
	err := testRepo.CreateToken(context.Background(), &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: "hash123",
		UserAgent: &ua,
		IPAddress: &ip,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	assert.NoError(t, err)
}

func TestGetToken_Success(t *testing.T) {
	truncateAll(t)
	user := createTestUser(t, "get-token@example.com", "gettokenuser")

	ua := "TestAgent"
	ip := "10.0.0.1"
	err := testRepo.CreateToken(context.Background(), &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: "findme-hash",
		UserAgent: &ua,
		IPAddress: &ip,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	token, err := testRepo.GetToken(context.Background(), "findme-hash")

	require.NoError(t, err)
	assert.Equal(t, user.ID, token.UserID)
	assert.Equal(t, "findme-hash", token.TokenHash)
	assert.Equal(t, domain.Human, token.Role) // from JOIN with auth.users
	assert.Nil(t, token.RevokedAt)
}

func TestGetToken_NotFound(t *testing.T) {
	truncateAll(t)

	_, err := testRepo.GetToken(context.Background(), "nonexistent-hash")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTokenNotFound)
}

func TestRevokeToken_Success(t *testing.T) {
	truncateAll(t)
	user := createTestUser(t, "revoke@example.com", "revokeuser")

	err := testRepo.CreateToken(context.Background(), &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: "revoke-hash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	err = testRepo.RevokeToken(context.Background(), "revoke-hash")
	assert.NoError(t, err)

	// Verify token is revoked
	token, err := testRepo.GetToken(context.Background(), "revoke-hash")
	require.NoError(t, err)
	assert.NotNil(t, token.RevokedAt)
}

func TestRevokeToken_Idempotent(t *testing.T) {
	truncateAll(t)
	user := createTestUser(t, "idem@example.com", "idemuser")

	err := testRepo.CreateToken(context.Background(), &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: "idem-hash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// Revoke twice — second call should not error
	err = testRepo.RevokeToken(context.Background(), "idem-hash")
	assert.NoError(t, err)

	err = testRepo.RevokeToken(context.Background(), "idem-hash")
	assert.NoError(t, err)
}

func TestRevokeAllTokens(t *testing.T) {
	truncateAll(t)
	user := createTestUser(t, "revokeall@example.com", "revokealluser")

	// Create 3 tokens
	for i, hash := range []string{"all-hash-1", "all-hash-2", "all-hash-3"} {
		_ = i
		err := testRepo.CreateToken(context.Background(), &domain.RefreshToken{
			UserID:    user.ID,
			TokenHash: hash,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		})
		require.NoError(t, err)
	}

	err := testRepo.RevokeAllTokens(context.Background(), user.ID)
	assert.NoError(t, err)

	// All tokens should be revoked
	for _, hash := range []string{"all-hash-1", "all-hash-2", "all-hash-3"} {
		token, err := testRepo.GetToken(context.Background(), hash)
		require.NoError(t, err)
		assert.NotNil(t, token.RevokedAt, "token %s should be revoked", hash)
	}
}

func TestCascadeDelete_UserRemovesTokens(t *testing.T) {
	truncateAll(t)
	user := createTestUser(t, "cascade@example.com", "cascadeuser")

	err := testRepo.CreateToken(context.Background(), &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: "cascade-hash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// Delete user directly via SQL
	testDB.MustExec("DELETE FROM auth.users WHERE id = $1", user.ID)

	// Token should be gone (CASCADE)
	_, err = testRepo.GetToken(context.Background(), "cascade-hash")
	assert.ErrorIs(t, err, domain.ErrTokenNotFound)
}

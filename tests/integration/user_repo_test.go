//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
)

type userRepo interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

func TestCreateUser_Success(t *testing.T) {
	truncateAll(t)

	user, err := testRepo.CreateUser(context.Background(), &domain.User{
		Email:        "alice@example.com",
		PasswordHash: "hashed",
		Username:     "alice",
		Role:         domain.Human,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.Equal(t, domain.Human, user.Role)
	assert.Equal(t, 1, user.TokenVersion)
	assert.False(t, user.CreatedAt.IsZero())
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	truncateAll(t)

	_, err := testRepo.CreateUser(context.Background(), &domain.User{
		Email:        "dup@example.com",
		PasswordHash: "hashed",
		Username:     "user1",
		Role:         domain.Human,
	})
	require.NoError(t, err)

	_, err = testRepo.CreateUser(context.Background(), &domain.User{
		Email:        "dup@example.com",
		PasswordHash: "hashed",
		Username:     "user2",
		Role:         domain.AI,
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	truncateAll(t)

	_, err := testRepo.CreateUser(context.Background(), &domain.User{
		Email:        "first@example.com",
		PasswordHash: "hashed",
		Username:     "samename",
		Role:         domain.Human,
	})
	require.NoError(t, err)

	_, err = testRepo.CreateUser(context.Background(), &domain.User{
		Email:        "second@example.com",
		PasswordHash: "hashed",
		Username:     "samename",
		Role:         domain.AI,
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}

func TestGetUserByEmail_Success(t *testing.T) {
	truncateAll(t)

	created, err := testRepo.CreateUser(context.Background(), &domain.User{
		Email:        "find@example.com",
		PasswordHash: "hashed",
		Username:     "findme",
		Role:         domain.AI,
	})
	require.NoError(t, err)

	found, err := testRepo.GetUserByEmail(context.Background(), "find@example.com")

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "find@example.com", found.Email)
	assert.Equal(t, domain.AI, found.Role)
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	truncateAll(t)

	_, err := testRepo.GetUserByEmail(context.Background(), "ghost@example.com")

	assert.Error(t, err)
}


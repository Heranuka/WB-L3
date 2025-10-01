package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"wb-l3.7/internal/config"
	"wb-l3.7/internal/domain"

	mock_auth "wb-l3.7/internal/service/auth/mocks"
	"wb-l3.7/pkg/jwt"
)

func TestAuth_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserStorage := mock_auth.NewMockUserStorage(ctrl)
	mockTokensStorage := mock_auth.NewMockTokensStorage(ctrl)

	a, err := NewAuth(testConfig(), mockUserStorage, mockTokensStorage)
	assert.NoError(t, err)

	ctx := context.Background()
	user := &domain.User{Nickname: "testuser", PasswordHash: "hashedpass"}

	mockUserStorage.EXPECT().
		SaveUser(ctx, gomock.Any()).
		Return(int64(123), nil).
		Times(1)

	err = a.Register(ctx, user.Nickname, "password", nil)
	assert.NoError(t, err)
}

func TestAuth_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserStorage := mock_auth.NewMockUserStorage(ctrl)
	mockTokensStorage := mock_auth.NewMockTokensStorage(ctrl)

	a, err := NewAuth(testConfig(), mockUserStorage, mockTokensStorage)
	assert.NoError(t, err)
	hash, _ := a.hasher.Hash("password")
	ctx := context.Background()
	fakeUser := &domain.User{
		ID:           1,
		Nickname:     "testuser",
		PasswordHash: hash,
	}

	mockUserStorage.EXPECT().
		GetUser(ctx, "testuser").
		Return(fakeUser, nil).
		AnyTimes()

	mockTokensStorage.EXPECT().
		StoreRefreshToken(ctx, fakeUser.ID, gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	// Call login with correct password
	tokens, user, err := a.Login(ctx, "testuser", "password")
	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.Equal(t, fakeUser.ID, user.ID)

	// Call login with incorrect password
	_, _, err = a.Login(ctx, "testuser", "wrongpassword")
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestAuth_GenerateTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserStorage := mock_auth.NewMockUserStorage(ctrl)
	mockTokensStorage := mock_auth.NewMockTokensStorage(ctrl)

	a, err := NewAuth(testConfig(), mockUserStorage, mockTokensStorage)
	assert.NoError(t, err)

	ctx := context.Background()
	user := &domain.User{ID: 1, Nickname: "testuser", Roles: []jwt.Role{}}

	mockTokensStorage.EXPECT().
		StoreRefreshToken(ctx, user.ID, gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	tokens, err := a.GenerateTokens(ctx, user)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.WithinDuration(t, time.Now(), tokens.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().Add(a.refreshTokenTTL), tokens.ExpiresAt, time.Second)
}

func TestAuth_Refresh(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserStorage := mock_auth.NewMockUserStorage(ctrl)
	mockTokensStorage := mock_auth.NewMockTokensStorage(ctrl)

	a, err := NewAuth(testConfig(), mockUserStorage, mockTokensStorage)
	assert.NoError(t, err)

	ctx := context.Background()
	user := &domain.User{ID: 1, Nickname: "testuser", Roles: []jwt.Role{}}
	refreshToken := "refresh-token"

	mockTokensStorage.EXPECT().
		GetUserByRefreshToken(ctx, refreshToken).
		Return(user, nil).
		Times(1)

	mockTokensStorage.EXPECT().
		StoreRefreshToken(ctx, user.ID, gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	tokens, err := a.Refresh(ctx, refreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, tokens)
}

func testConfig() config.Config {
	return config.Config{
		AuthConfig: config.AuthConfig{
			JWTSigningKey:   "secret",
			PasswordSalt:    "salt",
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: 24 * time.Hour,
		},
	}
}

// Helper method for hashing passwords in login test
func (h *Auth) HashStringForTest(password string) string {
	hash, _ := h.hasher.Hash(password)
	return hash
}

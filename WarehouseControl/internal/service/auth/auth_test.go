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

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserStorage := mock_auth.NewMockUserStorage(ctrl)

	a, err := NewAuth(config.Config{}, mockUserStorage, nil)
	assert.NoError(t, err)

	mockUserStorage.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(int64(1), nil)

	err = a.Register(context.Background(), "testuser", "password", []jwt.Role{jwt.Viewer})
	assert.NoError(t, err)
}

func TestLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserStorage := mock_auth.NewMockUserStorage(ctrl)
	mockTokensStorage := mock_auth.NewMockTokensStorage(ctrl)

	cfg := config.Config{
		AuthConfig: config.AuthConfig{
			JWTSigningKey:   "secret",
			PasswordSalt:    "salt",
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: time.Hour * 24,
		},
	}

	a, err := NewAuth(cfg, mockUserStorage, mockTokensStorage)
	assert.NoError(t, err)
	hash, _ := a.hasher.Hash("password")
	user := &domain.User{
		ID:           1,
		Nickname:     "testuser",
		Roles:        []jwt.Role{jwt.Viewer},
		PasswordHash: hash,
	}

	mockUserStorage.EXPECT().GetUser(gomock.Any(), "testuser").Return(user, nil)
	mockTokensStorage.EXPECT().StoreRefreshToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	tokens, userResp, err := a.Login(context.Background(), "testuser", "password")
	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotNil(t, userResp)
}

func TestRefresh_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserStorage := mock_auth.NewMockUserStorage(ctrl)
	mockTokensStorage := mock_auth.NewMockTokensStorage(ctrl)

	a, err := NewAuth(config.Config{}, mockUserStorage, mockTokensStorage)
	assert.NoError(t, err)

	user := &domain.User{
		ID:       1,
		Nickname: "user1",
		Roles:    []jwt.Role{jwt.Viewer},
	}

	mockTokensStorage.EXPECT().GetUserByRefreshToken(gomock.Any(), "refresh123").Return(user, nil)
	mockTokensStorage.EXPECT().StoreRefreshToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	tokens, err := a.Refresh(context.Background(), "refresh123")
	assert.NoError(t, err)
	assert.NotNil(t, tokens)
}

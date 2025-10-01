package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"wb-l3.7/internal/domain"
	mock_auth "wb-l3.7/internal/ports/rest/auth/mocks"
	"wb-l3.7/pkg/jwt"
)

// Тест регистрации успешный случай
func TestHandler_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	mockAuth := mock_auth.NewMockServiceAuth(ctrl)
	handler := NewHandler(logger, mockAuth) // или любой ваш zerolog.Logger

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	reqBody := `{"nickname":"testuser","password":"password","roles":["viewer"]}`
	req := httptest.NewRequest("POST", "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	mockAuth.EXPECT().Register(gomock.Any(), "testuser", "password", gomock.Any()).Return(nil)

	handler.Register(ctx)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, strings.ToLower(w.Body.String()), "success")
}

// Тест регистрации - ошибка дублирования никнейма
func TestHandler_Register_NicknameExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	mockAuth := mock_auth.NewMockServiceAuth(ctrl)
	handler := NewHandler(logger, mockAuth)

	mockAuth.EXPECT().
		Register(gomock.Any(), "existinguser", "pass", gomock.Any()).
		Return(domain.ErrNicknameAlreadyExist)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	body := `{"nickname":"existinguser","password":"pass","roles":["viewer"]}`
	req := httptest.NewRequest("POST", "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx.Request = req
	handler.Register(ctx)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, strings.ToLower(w.Body.String()), "nickname already exist")
}

// Тест логина успешный случай
func TestHandler_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	mockAuth := mock_auth.NewMockServiceAuth(ctrl)
	handler := NewHandler(logger, mockAuth)

	tokens := &domain.Tokens{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		CreatedAt:    time.Unix(123, 0), // конвертация int64 в time.Time
		ExpiresAt:    time.Unix(456, 0),
	}
	user := &domain.User{
		ID:       1,
		Nickname: "testuser",
		Roles:    []jwt.Role{jwt.Viewer},
	}

	mockAuth.EXPECT().
		Login(gomock.Any(), "testuser", "password").
		Return(tokens, user, nil)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	body := `{"nickname":"testuser","password":"password"}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx.Request = req
	handler.Login(ctx)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, strings.ToLower(w.Body.String()), "access-token")
}

// Тест логина - неверные данные
func TestHandler_Login_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	mockAuth := mock_auth.NewMockServiceAuth(ctrl)
	handler := NewHandler(logger, mockAuth)

	mockAuth.EXPECT().
		Login(gomock.Any(), "baduser", "badpass").
		Return(nil, nil, domain.ErrInvalidCredentials)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	body := `{"nickname":"baduser","password":"badpass"}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx.Request = req
	handler.Login(ctx)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, strings.ToLower(w.Body.String()), "invalid credentials")
}

// Тест обновления токена успешный
func TestHandler_RefreshToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	mockAuth := mock_auth.NewMockServiceAuth(ctrl)
	handler := NewHandler(logger, mockAuth)

	tokens := &domain.Tokens{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}

	mockAuth.EXPECT().
		Refresh(gomock.Any(), "old-refresh-token").
		Return(tokens, nil)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	body := `{"refresh_token":"old-refresh-token"}`
	req := httptest.NewRequest("POST", "/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx.Request = req
	handler.RefreshToken(ctx)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, strings.ToLower(w.Body.String()), "new-access-token")
}

// Тест обновления токена - пользователь не найден
func TestHandler_RefreshToken_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	mockAuth := mock_auth.NewMockServiceAuth(ctrl)
	handler := NewHandler(logger, mockAuth)

	mockAuth.EXPECT().
		Refresh(gomock.Any(), "bad-token").
		Return(nil, domain.ErrUserNotFound)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	body := `{"refresh_token":"bad-token"}`
	req := httptest.NewRequest("POST", "/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx.Request = req
	handler.RefreshToken(ctx)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, strings.ToLower(w.Body.String()), "user not found")
}

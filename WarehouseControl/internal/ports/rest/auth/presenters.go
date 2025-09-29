package auth

import (
	"time"

	"wb-l3.7/pkg/jwt"
)

type registerRequest struct {
	Nickname string     `json:"nickname" validate:"required"`
	Password string     `json:"password" validate:"required"`
	Roles    []jwt.Role `json:"roles"`
}

// Запрос логина пользователя
type loginRequest struct {
	Nickname string `json:"nickname" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Ответ с токенами после аутентификации
type tokenResponse struct {
	Nickname     string     `json:"nickname"`
	UserID       int64      `json:"user_id"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	Roles        []jwt.Role `json:"roles"`
}

// Запрос обновления токена
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

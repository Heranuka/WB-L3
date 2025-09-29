package domain

import (
	"time"

	"wb-l3.7/pkg/jwt"
)

type User struct {
	ID           int64      `json:"id"`
	Nickname     string     `json:"nickname"`
	PasswordHash string     `json:"password_hash"`
	Roles        []jwt.Role `json:"roles"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Tokens struct {
	ID           int64     // bigint (bigserial) в базе
	UserID       int64     `json:"user_id"`
	AccessToken  string    `json:"access-token"`
	RefreshToken string    `json:"refresh-token"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Session struct {
	RefreshToken string    `json:"refresh_token"`
	AccessToken  string    `json:"access_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Room struct {
	ID        int64 // bigint (bigserial) для ID
	Name      string
	CreatedAt time.Time `json:"created_at"`
}

type UserClaims struct {
	UserID   int64      `json:"user_id"` // bigint
	Nickname string     `json:"nickname"`
	Roles    []jwt.Role `json:"roles"`
}

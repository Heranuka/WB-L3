package auth

import (
	"context"
	"fmt"
	"time"

	"wb-l3.7/internal/config"
	"wb-l3.7/internal/domain"
	"wb-l3.7/pkg/hash"
	"wb-l3.7/pkg/jwt"
)

type UserStorage interface {
	SaveUser(ctx context.Context, user *domain.User) (int64, error)
	GetUser(ctx context.Context, nickname string) (*domain.User, error)
	SetSession(ctx context.Context, userID int64, session *domain.Session) error
	GetBySession(ctx context.Context, refreshToken string) (*domain.User, error)
}

type Auth struct {
	storage         UserStorage
	tokenManager    jwt.TokenManager
	hasher          hash.PasswordHasher
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuth(cfg config.Config, storage UserStorage) (*Auth, error) {
	tokenManager, err := jwt.NewManager(cfg.AuthConfig.JWTSigningKey)
	if err != nil {
		return nil, err
	}

	hasher, err := hash.NewSHa1Hasher(cfg.AuthConfig.PasswordSalt)
	if err != nil {
		return nil, err
	}

	return &Auth{
		storage:         storage,
		hasher:          hasher,
		tokenManager:    tokenManager,
		accessTokenTTL:  cfg.AuthConfig.AccessTokenTTL,
		refreshTokenTTL: cfg.AuthConfig.RefreshTokenTTL,
	}, nil
}

func (a *Auth) Register(ctx context.Context, nickname, password string, roles []jwt.Role) error {
	passHash, err := a.hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("service.Auth.Register: %w", err)
	}
	user := &domain.User{
		Nickname:     nickname,
		PasswordHash: passHash,
		Roles:        roles,
	}
	_, err = a.storage.SaveUser(ctx, user)
	if err != nil {
		return fmt.Errorf("service.Auth.Register: %w", err)
	}

	return nil
}

func (a *Auth) Login(ctx context.Context, nickname, password string) (*domain.Tokens, *domain.User, error) {
	passwordHash, err := a.hasher.Hash(password)
	if err != nil {
		return nil, nil, fmt.Errorf("service.Auth.Login.Hash: %w", err)
	}
	user, err := a.storage.GetUser(ctx, nickname)
	if err != nil {
		return nil, nil, fmt.Errorf("service.Auth.Login.GetUser: %w", err)
	}

	if user.PasswordHash != passwordHash {
		return nil, nil, domain.ErrInvalidCredentials
	}

	session, err := a.CreateSession(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("service.Auth.Login.CreateSession: %w", err)
	}
	return session, user, nil
}

func (a *Auth) CreateSession(ctx context.Context, user *domain.User) (*domain.Tokens, error) {
	// --- ИЗМЕНЕНИЕ ЗДЕСЬ ---
	// Передаем user.Roles в NewAccessToken
	accessToken, err := a.tokenManager.NewJWT(user.ID, user.Nickname, user.Roles, a.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("service.Auth.CreateSession.accessToken: %w", err)
	}
	// ... остальной код без изменений ...
	refreshToken, err := a.tokenManager.NewRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("service.Auth.CreateSession.refreshToken: %w", err)
	}

	tokens := &domain.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	session := &domain.Session{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		ExpiresAt:    time.Now().UTC().Add(a.refreshTokenTTL),
	}

	err = a.storage.SetSession(ctx, user.ID, session)
	if err != nil {
		return nil, fmt.Errorf("service.Auth.CreateSession.SetSession: %w", err)
	}

	return tokens, nil
}
func (a *Auth) Refresh(ctx context.Context, token string) (*domain.Tokens, error) {
	user, err := a.storage.GetBySession(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("service.Auth.Refresh.GetBySession: %w", err)
	}

	return a.CreateSession(ctx, user)
}

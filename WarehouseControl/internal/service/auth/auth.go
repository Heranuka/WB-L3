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

//go:generate mockgen -source=auth.go -destination=mocks/mock.go
type UserStorage interface {
	SaveUser(ctx context.Context, user *domain.User) (int64, error)
	GetUser(ctx context.Context, nickname string) (*domain.User, error)
}

type TokensStorage interface {
	StoreRefreshToken(ctx context.Context, userID int64, refreshToken string, expiresAt string) error
	GetUserByRefreshToken(ctx context.Context, refreshToken string) (*domain.User, error)
}

type Auth struct {
	storage         UserStorage
	token           TokensStorage
	tokenManager    jwt.TokenManager
	hasher          hash.PasswordHasher
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuth(cfg config.Config, userStorage UserStorage, tokenStorage TokensStorage) (*Auth, error) {
	tokenManager, err := jwt.NewManager(cfg.AuthConfig.JWTSigningKey)
	if err != nil {
		return nil, err
	}

	hasher, err := hash.NewSHa1Hasher(cfg.AuthConfig.PasswordSalt)
	if err != nil {
		return nil, err
	}

	return &Auth{
		storage:         userStorage,
		token:           tokenStorage,
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

	tokens, err := a.GenerateTokens(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("service.Auth.Login.CreateSession: %w", err)
	}
	return tokens, user, nil
}

func (a *Auth) GenerateTokens(ctx context.Context, user *domain.User) (*domain.Tokens, error) {
	accessToken, err := a.tokenManager.NewJWT(user.ID, user.Nickname, user.Roles, a.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("service.Auth.CreateSession.accessToken: %w", err)
	}
	// ... остальной код без изменений ...
	refreshToken, err := a.tokenManager.NewRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("service.Auth.CreateSession.refreshToken: %w", err)
	}
	expiresAt := time.Now().Add(a.refreshTokenTTL).Format(time.RFC3339)
	err = a.token.StoreRefreshToken(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("service.Auth.CreateSession.storeRefreshToken: %w", err)
	}

	tokens := &domain.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return tokens, nil
}

func (a *Auth) Refresh(ctx context.Context, token string) (*domain.Tokens, error) {
	user, err := a.token.GetUserByRefreshToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("service.Auth.Refresh.GetUserByRefreshToken: %w", err)
	}

	return a.GenerateTokens(ctx, user)
}

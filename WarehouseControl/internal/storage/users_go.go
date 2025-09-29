package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"wb-l3.7/internal/domain"
)

func (pg *Postgres) SaveUser(ctx context.Context, user *domain.User) (int64, error) {
	row := pg.pool.QueryRow(ctx, "INSERT INTO users(nickname, password_hash, roles) VALUES ($1, $2, $3) RETURNING id", user.Nickname, user.PasswordHash, user.Roles)
	//tokenRow := pg.pool.QueryRow(ctx, "INSERT INTO tokens(user_id, refresh_token, access_token, created_at, expires_at) VALUES ($1, $2, $3, $4, $5)")
	var id int64
	if err := row.Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName != "" {
			return -1, domain.ErrNicknameAlreadyExist
		}

		return -1, fmt.Errorf("storage.pg.SaveUser: %w", err)
	}

	return id, nil
}

func (pg *Postgres) GetUser(ctx context.Context, nickname string) (*domain.User, error) {
	row := pg.pool.QueryRow(ctx, "SELECT id, nickname, password_hash, roles FROM users WHERE nickname = $1", nickname)

	var user domain.User
	err := row.Scan(&user.ID, &user.Nickname, &user.PasswordHash, &user.Roles)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("storage.pg.GetUser: %w", err)
	}

	return &user, nil
}

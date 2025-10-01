package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"

	"wb-l3.7/internal/domain"
)

type User interface {
	SaveUser(ctx context.Context, user *domain.User) (int64, error)
	GetUser(ctx context.Context, nickname string) (*domain.User, error)
}

func (pg *Postgres) SaveUser(ctx context.Context, user *domain.User) (int64, error) {
	rolesJSON, err := json.Marshal(user.Roles)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal roles: %w", err)
	}

	row := pg.db.Master.QueryRowContext(ctx, "INSERT INTO users(nickname, password_hash, roles) VALUES ($1, $2, $3) RETURNING id", user.Nickname, user.PasswordHash, rolesJSON)
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
	row := pg.db.Master.QueryRowContext(ctx, "SELECT id, nickname, password_hash, roles, created_at, updated_at FROM users WHERE nickname = $1", nickname)

	var user domain.User
	var rolesJSON []byte
	err := row.Scan(&user.ID, &user.Nickname, &user.PasswordHash, &rolesJSON, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("storage.pg.GetUser: %w", err)
	}

	err = json.Unmarshal(rolesJSON, &user.Roles)
	if err != nil {
		// handle error
	}
	return &user, nil
}

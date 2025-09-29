package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"wb-l3.7/internal/domain"
)

func (pg *Postgres) SetSession(ctx context.Context, userID int64, session *domain.Session) error {
	_, err := pg.pool.Exec(ctx, `
        INSERT INTO tokens (user_id, refresh_token, access_token, created_at, expires_at)
        VALUES ($1, $2, $3, NOW(), $4)
        ON CONFLICT (user_id) DO UPDATE
        SET refresh_token = EXCLUDED.refresh_token,
            access_token = EXCLUDED.access_token,
            expires_at = EXCLUDED.expires_at
    `, userID, session.RefreshToken, session.AccessToken, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("storage.pg.SetSession: %w", err)
	}

	return nil
}

func (pg *Postgres) GetBySession(ctx context.Context, refreshToken string) (*domain.User, error) {
	row := pg.pool.QueryRow(ctx, `SELECT u.id, u.nickname, u.roles
    FROM tokens t
    JOIN users u ON u.id = t.user_id
    WHERE t.refresh_token = $1`, refreshToken)

	var user domain.User
	err := row.Scan(&user.ID, &user.Nickname, &user.Roles)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("storage.pg.GetBySession: %w", err)
	}

	return &user, nil
}

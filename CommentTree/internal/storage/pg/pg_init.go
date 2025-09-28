package pg

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	log  *slog.Logger
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, log *slog.Logger) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(os.Getenv("DB_URL"))
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Postgres: %w", err)
	}
	return &Postgres{
		log:  log,
		pool: pool,
	}, nil
}

func (s *Postgres) Stop() {
	s.pool.Close()
}

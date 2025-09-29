package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"wb-l3.7/internal/config"
	"wb-l3.7/pkg/e"
)

type Postgres struct {
	logger *slog.Logger
	pool   *pgxpool.Pool
}

func NewPostgres(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Postgres, error) {
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
		cfg.Postgres.SSLMode,
	)
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, e.Wrap("storage.pg.NewPostgres.ParseConfig", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		return nil, e.Wrap("storage.pg.NewPostgres.NewWithConfig", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, e.Wrap("storage.pg.NewPostgres.Ping", err)
	}

	return &Postgres{
		logger: logger,
		pool:   pool,
	}, nil
}

func (p *Postgres) CloseConnection() {
	p.pool.Close()
	stat := p.pool.Stat()
	if stat.AcquiredConns() > 0 {
		p.logger.Warn("postgres connections not fully closed after Close()", slog.Any("acquired connections", stat.AcquiredConns()))

	}
}

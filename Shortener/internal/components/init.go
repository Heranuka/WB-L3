package components

import (
	"context"
	"fmt"
	"shortener/internal/config"
	"shortener/internal/ports"
	"shortener/internal/service/cache"
	url_shortener "shortener/internal/service/linksAuth"
	"shortener/internal/service/render"
	"shortener/internal/storage/pg"

	"github.com/rs/zerolog"
)

type Components struct {
	logger     zerolog.Logger
	HttpServer *ports.Server
	Postgres   *pg.Postgres
	Redis      *cache.RedisService
}

func InitComponents(ctx context.Context, cfg *config.Config, logger zerolog.Logger) (*Components, error) {
	postgres, err := pg.NewPostgres(ctx, cfg)
	if err != nil {
		return nil, err
	}

	redis := cache.NewRedisService(cfg)

	render := render.New("./templates", logger)
	serviceURLShortener := url_shortener.NewURLShortener(logger, postgres, postgres, redis)

	httpServer := ports.NewServer(ctx, cfg, logger, serviceURLShortener, serviceURLShortener, render)

	return &Components{
		logger:     logger,
		HttpServer: httpServer,
		Postgres:   postgres,
		Redis:      redis,
	}, nil
}
func (c *Components) ShutdownAll() error {
	if err := c.HttpServer.Stop(); err != nil {
		c.logger.Error().
			Err(err).
			Msg("components.ShutdownAll: failed to shutdown http server")
		return fmt.Errorf("failed to close HTTP Server connection")
	}
	if err := c.Postgres.Close(); err != nil {
		c.logger.Error().
			Err(err).
			Msg("components.ShutdownAll: failed to stop postgres")
		return fmt.Errorf("failed to close postgres connection")
	}
	return nil
}

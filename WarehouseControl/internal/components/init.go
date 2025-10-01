package components

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"wb-l3.7/internal/config"
	"wb-l3.7/internal/ports"
	"wb-l3.7/internal/service/auth"
	"wb-l3.7/internal/service/items"
	"wb-l3.7/internal/service/render"
	"wb-l3.7/internal/storage"
	"wb-l3.7/pkg/jwt"
)

type Components struct {
	logger     zerolog.Logger
	postgres   *storage.Postgres
	HttpServer *ports.Server
}

func InitComponents(ctx context.Context, logger zerolog.Logger, cfg *config.Config) (*Components, error) {
	logger.Info().Msg("Starting components initialization")

	pg, err := storage.NewPostgres(ctx, cfg)
	if err != nil {
		logger.Error().Err(err).Msg("Postgres initialization failed")
		return nil, fmt.Errorf("components.init.InitComponents.postgres failed: %w", err)
	}
	logger.Info().Msg("Postgres initialized")

	tokenManager, err := jwt.NewManager(cfg.AuthConfig.JWTSigningKey)
	if err != nil {
		logger.Error().Err(err).Msg("TokenManager initialization failed")
		return nil, err
	}
	logger.Info().Msg("TokenManager initialized")

	cwd, err := os.Getwd()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get current directory")
		return nil, err
	}
	logger.Info().Str("cwd", cwd).Msg("Current working directory obtained")

	renderService := render.New(cwd+"/templates", logger)
	logger.Info().Msg("Render service initialized")

	serviceItem := items.NewItemService(pg, pg)
	serviceAuth, err := auth.NewAuth(*cfg, pg, pg)
	if err != nil {
		logger.Error().Err(err).Msg("Auth service initialization failed")
		return nil, err
	}
	logger.Info().Msg("Auth service initialized")
	logger.Info().Msg("Item service initialized")

	httpServer, err := ports.NewServer(cfg, serviceAuth, serviceItem, serviceItem, renderService, logger, tokenManager)
	if err != nil {
		logger.Error().Err(err).Msg("HTTP server initialization failed")
		return nil, err
	}
	logger.Info().Msg("HTTP server initialized")

	logger.Info().Msg("All components initialized successfully")

	return &Components{
		logger:     logger,
		postgres:   pg,
		HttpServer: httpServer,
	}, nil
}

func (c *Components) Shutdown() error {
	c.logger.Info().Msg("Starting shutdown of components")
	if err := c.postgres.Close(); err != nil {
		return err
	}

	c.logger.Info().Msg("Postgres connection closed")

	c.HttpServer.Stop()
	c.logger.Info().Msg("HTTP server stopped")

	c.logger.Info().Msg("Components shutdown completed successfully")
	return nil
}

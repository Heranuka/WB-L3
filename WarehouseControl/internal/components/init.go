package components

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"wb-l3.7/internal/config"
	"wb-l3.7/internal/ports"
	"wb-l3.7/internal/service/auth"
	"wb-l3.7/internal/service/items"
	"wb-l3.7/internal/service/render"
	"wb-l3.7/internal/storage"
	"wb-l3.7/pkg/jwt"
	"wb-l3.7/pkg/logger"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Components struct {
	logger     *slog.Logger
	postgres   *storage.Postgres
	HttpServer *ports.Server
}

func InitComponents(ctx context.Context, logger *slog.Logger, cfg *config.Config) (*Components, error) {
	pg, err := storage.NewPostgres(ctx, *cfg, logger)
	if err != nil {
		logger.Error("postgres error", "error", err.Error())
		return nil, fmt.Errorf("components.init.InitComponents.postgres failed: %w", err)
	}
	tokenManager, err := jwt.NewManager(cfg.AuthConfig.JWTSigningKey)
	if err != nil {
		return nil, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("failed to get current directory: %v", err)
		return nil, err
	}
	fmt.Println("Current work directory:", cwd)

	render := render.New(cwd+"/templates", logger)

	serviceItem := items.NewItemService(pg, pg)
	serviceAuth, err := auth.NewAuth(*cfg, pg)
	if err != nil {
		return nil, err
	}

	httpServer, err := ports.NewServer(cfg, serviceAuth, serviceItem, serviceItem, *render, logger, tokenManager)
	if err != nil {
		return nil, err
	}

	return &Components{
		logger:     logger,
		postgres:   pg,
		HttpServer: httpServer,
	}, nil
}

func (c *Components) Shutdown() error {
	var errs []error
	c.postgres.CloseConnection()
	c.HttpServer.Stop()

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = logger.SetupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:

		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

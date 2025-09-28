package components

import (
	"context"
	"log/slog"
	"os"
	"test_18/internal/config"
	"test_18/internal/ports"
	"test_18/internal/service"
	"test_18/internal/storage/pg"

	"test_18/pkg/e"
	"test_18/pkg/logger"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Components struct {
	HttpServer *ports.Server
	postgres   *pg.Postgres
	logger     *slog.Logger
}

func InitComponents(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Components, error) {

	pg, err := pg.NewPostgres(ctx, cfg, logger)
	if err != nil {
		return nil, e.Wrap("components.init.InitComponents", err)
	}
	eventService := service.NewService(logger, pg)

	httpServer := ports.NewServer(ctx, &cfg, logger, *eventService)

	return &Components{
		postgres:   pg,
		HttpServer: httpServer,
		logger:     logger,
	}, nil
}

func (c *Components) ShutDownComponents() error {
	c.postgres.CloseDB()
	if err := c.HttpServer.StopHTTPServer(); err != nil {
		c.logger.Error("Failed to shutDown Server", slog.String("error", err.Error()))
		return e.Wrap("components.ShutDownComponents", err)
	}
	return nil
}

func SetupLogger(cfg string) *slog.Logger {
	var log *slog.Logger

	switch envLocal {
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

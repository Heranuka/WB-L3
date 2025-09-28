package components

import (
	"commentTree/internal/config"
	"commentTree/internal/ports"
	"commentTree/internal/service"
	"net/http"

	"commentTree/internal/storage/pg"
	slogpretty "commentTree/pkg/logger"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Components struct {
	HttpServer *ports.Server
	pg         *pg.Postgres
}
type DummyRenderService struct{}

func (d *DummyRenderService) Home(w http.ResponseWriter) {
	// Например, отдать пустую страницу или редирект
	w.Write([]byte("Hello Home Page"))
}

func InitComponents(ctx context.Context, logger *slog.Logger, cfg *config.Config) (*Components, error) {
	pg, err := pg.NewPostgres(ctx, logger)
	if err != nil {
		logger.Error("Failed to init Postgres", slog.String("error", err.Error()))
		return nil, err
	}

	services := service.NewService(logger, pg)
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("failed to get current directory: %v", err)
		return nil, err
	}
	fmt.Println("Current work directory:", cwd)
	dummyService := &DummyRenderService{}
	render := service.NewRender(cwd, dummyService, logger)

	server := ports.NewServer(ctx, logger, cfg, *services, *render)
	return &Components{
		HttpServer: server,
		pg:         pg,
	}, nil
}

func (c *Components) StopComponents() error {
	if err := c.HttpServer.Stop(); err != nil {
		return err
	}

	c.pg.Stop()

	return nil
}

func SetupLogger() *slog.Logger {
	var logger *slog.Logger
	var env = "local"
	switch env {
	case envLocal:
		logger = slogpretty.SetupPrettySlog()
	case envDev:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:

		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		// Теперь, если env неизвестен, мы вернем этот дефолтный логгер.
	}

	return logger
}

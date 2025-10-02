package components

import (
	"commentTree/internal/config"
	"commentTree/internal/ports"
	"commentTree/internal/service"
	"net/http"

	"commentTree/internal/storage/pg"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/rs/zerolog"
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

func InitComponents(ctx context.Context, logger zerolog.Logger, cfg *config.Config) (*Components, error) {
	pg, err := pg.NewPostgres(ctx, cfg)
	if err != nil {
		logger.Error().Err(err).Msg("Postgres initialization failed")
		return nil, fmt.Errorf("components.init.InitComponents.postgres failed: %w", err)
	}
	logger.Info().Msg("Postgres service initialized")

	services := service.NewService(logger, pg)
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("failed to get current directory: %v", err)
		return nil, err
	}
	fmt.Println("Current work directory:", cwd)

	renderService := service.NewRender(cwd+"/templates", logger)
	logger.Info().Msg("Render service initialized")

	server := ports.NewServer(ctx, logger, cfg, services, renderService)
	return &Components{
		HttpServer: server,
		pg:         pg,
	}, nil
}

func (c *Components) StopComponents() error {
	if err := c.HttpServer.Stop(); err != nil {
		return err
	}

	if err := c.pg.Close(); err != nil {
		return err
	}

	return nil
}

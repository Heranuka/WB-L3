package components

import (
	"context"
	"delay/internal/broker"
	"delay/internal/config"
	"delay/internal/ports"
	"delay/internal/service"
	"delay/internal/service/cache"
	"delay/internal/service/notificationService"
	"delay/internal/service/render"
	postgres "delay/internal/storage/postrgres"
	"fmt"
	"log"
	"os"

	"github.com/rs/zerolog"
)

type Components struct {
	logger     zerolog.Logger
	HttpServer *ports.Server
	Postgres   *postgres.Postgres
	Redis      *cache.RedisService
	Broker     *broker.RabbitMQ
}

func InitComponents(ctx context.Context, cfg *config.Config, logger zerolog.Logger) (*Components, error) {
	postgres, err := postgres.NewPostgres(ctx, cfg)
	if err != nil {
		return nil, err
	}

	rabbitMq, err := broker.NewRabbitMQ(ctx, logger, cfg, "notifications.delayed", "notifications.dlx", "notifications.dlq", nil)
	if err != nil {
		logger.Error().Err(err).Msg("Components: Rabbit failed")
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("failed to get current directory: %v", err)
		return nil, err
	}
	fmt.Println("Current work directory:", cwd)

	render := render.New(cwd+"/templates", logger)

	redis := cache.NewRedisService(cfg)
	service := service.NewService(*cfg, logger, postgres, redis, rabbitMq)

	emailChannel := notificationService.NewEmailChannel(cfg.EmailSmpt.SmptPort, cfg.EmailSmpt.SmptServer, cfg.EmailSmpt.SmptEmail, cfg.EmailSmpt.SmptPassword)
	telegramChannel, err := notificationService.NewTelegramChannel(cfg.TelegBot.Key, cfg.TelegBot.ChatID)
	if err != nil {
		return nil, err
	}
	if err := rabbitMq.Consume(ctx, emailChannel, telegramChannel); err != nil {
		logger.Error().Err(err).Msg("Failed to consume notifications")
		return nil, err
	}

	httpServer := ports.NewServer(ctx, cfg, logger, service, render)
	return &Components{
		logger:     logger,
		HttpServer: httpServer,
		Postgres:   postgres,
		Redis:      redis,
		Broker:     rabbitMq,
	}, nil
}

func (c *Components) ShutdownAll() error {
	c.logger.Info().Msg("ShutdownAll: starting...")
	c.Postgres.Close()
	err := c.Broker.Stop()
	if err != nil {
		return fmt.Errorf("could not properly stop Broker: %w", err)
	}

	c.logger.Info().Msg("ShutdownAll: finished")
	return nil
}

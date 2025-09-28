package service

import (
	"context"
	"delay/internal/broker"
	"delay/internal/config"
	"delay/internal/domain"
	"delay/internal/service/notificationService"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Интерфейс кеш-сервиса
type CacheService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
}

type Rabbit interface {
	UpdateStatus(ctx context.Context, noteID uuid.UUID, status string) error
	Consume(ctx context.Context, emailChannel, telegramChannel notificationService.NotificationChannel) error
	Publish(note *domain.Notification) error
}

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type NotificationService interface {
	Create(ctx context.Context, notification *domain.Notification) (uuid.UUID, error)
	Status(ctx context.Context, noteID uuid.UUID) (string, error)
	Cancel(ctx context.Context, noteID uuid.UUID) error
	GetAll(ctx context.Context) (*[]domain.Notification, error)
	Get(ctx context.Context, noteID uuid.UUID) (*domain.Notification, error)
	UpdateStatus(ctx context.Context, noteID uuid.UUID, status string) error
}

type Service struct {
	notificationService NotificationService
	cacheService        CacheService
	Config              *config.Config
	mu                  *sync.Mutex
	logger              zerolog.Logger
	rabbit              Rabbit
}

func NewService(cfg config.Config, logger zerolog.Logger, notificationService NotificationService, cacheService CacheService, rabbitmq broker.Rabbit) *Service {
	return &Service{
		notificationService: notificationService,
		cacheService:        cacheService,
		Config:              &cfg,
		mu:                  &sync.Mutex{},
		logger:              logger,
		rabbit:              rabbitmq,
	}
}

func (s *Service) Create(ctx context.Context, notification *domain.Notification) (uuid.UUID, error) {
	noteID, err := s.notificationService.Create(ctx, notification)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create notification in repo")
		return uuid.Nil, err
	}

	if err := s.notificationService.UpdateStatus(ctx, noteID, "created"); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update status to created")
	}

	note, err := s.Get(ctx, noteID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get notification")
	}

	s.logger.Info().
		Str("note_id", noteID.String()).
		Str("message", notification.Message).
		Msg("Notification created in repo")

	// Сохраняем статус в кэш
	cacheKey := "notification:status:" + noteID.String()
	if err := s.cacheService.Set(ctx, cacheKey, "created", 5*time.Minute); err != nil {
		s.logger.Error().Err(err).Msg("Failed to set cache for notification status")
	}
	// Публикуем в RabbitMQ
	if err := s.rabbit.Publish(note); err != nil {
		s.logger.Error().Err(err).Msg("Failed to publish notification to RabbitMQ")
		return uuid.Nil, err
	}

	s.logger.Info().
		Str("note_id", noteID.String()).
		Msg("Notification published to RabbitMQ")

	return noteID, nil
}

func (s *Service) Status(ctx context.Context, noteID uuid.UUID) (string, error) {
	cacheKey := "notification:status:" + noteID.String()
	var status string

	err := s.cacheService.Get(ctx, cacheKey, &status)
	if err == nil {
		s.logger.Debug().Str("noteID", noteID.String()).Msg("Cache hit for status")
		return status, nil
	}

	status, err = s.notificationService.Status(ctx, noteID)
	if err != nil {
		s.logger.Error().Err(err).Str("noteID", noteID.String()).Msg("Status not found")
		return "", fmt.Errorf("status not found")
	}

	if err := s.cacheService.Set(ctx, cacheKey, status, 5*time.Minute); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update cache for notification status")
	}

	return status, nil
}

func (s *Service) Cancel(ctx context.Context, noteID uuid.UUID) error {
	err := s.notificationService.Cancel(ctx, noteID)
	if err != nil {
		s.logger.Error().Err(err).Str("noteID", noteID.String()).Msg("Failed to delete notification")
		return err
	}

	cacheKey := "notification:status:" + noteID.String()
	if err := s.cacheService.Set(ctx, cacheKey, "", 0); err != nil {
		s.logger.Error().Err(err).Msg("Failed to clear cache for notification status")
	}

	s.logger.Info().Str("noteID", noteID.String()).Msg("Notification deleted")
	return nil
}

func (s *Service) GetAll(ctx context.Context) (*[]domain.Notification, error) {
	cacheKey := "notification:getall"
	var notifications []domain.Notification

	if err := s.cacheService.Get(ctx, cacheKey, &notifications); err == nil {
		s.logger.Debug().Msg("Cache hit for GetAll notifications")
		return &notifications, nil
	}

	notificationsPtr, err := s.notificationService.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	notifications = *notificationsPtr
	bytes, err := json.Marshal(notifications)
	if err == nil {
		if err := s.cacheService.Set(ctx, cacheKey, string(bytes), 5*time.Minute); err != nil {
			s.logger.Error().Err(err).Msg("Failed to set cache for GetAll notifications")
		}
	} else {
		s.logger.Error().Err(err).Msg("Failed to marshal notifications for caching")
	}

	return &notifications, nil
}

func (s *Service) Get(ctx context.Context, noteID uuid.UUID) (*domain.Notification, error) {
	return s.notificationService.Get(ctx, noteID)
}
func (s *Service) UpdateStatus(ctx context.Context, noteID uuid.UUID, status string) error {
	return s.notificationService.UpdateStatus(ctx, noteID, status)
}

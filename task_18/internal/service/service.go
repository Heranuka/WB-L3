package service

import (
	"context"
	"log/slog"
	"test_18/internal/domain"
	"time"
)

type DB interface {
	CreateEvent(ctx context.Context, event domain.Event) (int, error)
	UpdateEvent(ctx context.Context, id int, req domain.UpdateEventRequest) error
	DeleteEvent(ctx context.Context, id int) error
	GetEventsForDay(ctx context.Context, userID int, date time.Time) (domain.User, error)
	GetEventsForWeek(ctx context.Context, userID int, date time.Time) (domain.User, error)
	GetEventsForMonth(ctx context.Context, userID int, date time.Time) (domain.User, error)
	CreateUser(ctx context.Context) (int, error)
}

type Service struct {
	logger *slog.Logger
	db     DB
}

func NewService(logger *slog.Logger, db DB) *Service {
	return &Service{
		logger: logger,
		db:     db,
	}
}

func (s *Service) CreateEvent(ctx context.Context, event domain.Event) (int, error) {
	return s.db.CreateEvent(ctx, event)
}
func (s *Service) UpdateEvent(ctx context.Context, id int, req domain.UpdateEventRequest) error {
	return s.db.UpdateEvent(ctx, id, req)
}
func (s *Service) DeleteEvent(ctx context.Context, id int) error {
	return s.db.DeleteEvent(ctx, id)
}
func (s *Service) GetEventsForDay(ctx context.Context, userID int, date time.Time) (domain.User, error) {
	return s.db.GetEventsForDay(ctx, userID, date)
}
func (s *Service) GetEventsForWeek(ctx context.Context, userID int, date time.Time) (domain.User, error) {
	return s.db.GetEventsForWeek(ctx, userID, date)
}
func (s *Service) GetEventsForMonth(ctx context.Context, userID int, date time.Time) (domain.User, error) {
	return s.db.GetEventsForMonth(ctx, userID, date)
}

func (s *Service) CreateUser(ctx context.Context) (int, error) {
	return s.db.CreateUser(ctx)
}

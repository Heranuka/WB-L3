package service

import (
	"commentTree/internal/domain"
	"context"
	"log/slog"
)

type CommentService interface {
	Create(ctx context.Context, comment *domain.Comment) (int, error)
	GetById(ctx context.Context, id int) (*domain.Comment, error)
	Delete(ctx context.Context, id int) error
}

type Service struct {
	logger         *slog.Logger
	commentService CommentService
}

func NewService(logger *slog.Logger, commentService CommentService) *Service {
	return &Service{
		logger:         logger,
		commentService: commentService,
	}
}
func (s *Service) Create(ctx context.Context, comment *domain.Comment) (int, error) {
	return s.commentService.Create(ctx, comment)
}
func (s *Service) GetById(ctx context.Context, id int) (*domain.Comment, error) {
	return s.commentService.GetById(ctx, id)
}
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.commentService.Delete(ctx, id)
}

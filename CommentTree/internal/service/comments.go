package service

import (
	"commentTree/internal/domain"
	"context"

	"github.com/rs/zerolog"
)

//go:generate mockgen -source=comments.go -destination=mocks/mock.go
type CommentService interface {
	Create(ctx context.Context, comment *domain.Comment) (int, error)
	Delete(ctx context.Context, id int) error
	GetRootComments(ctx context.Context, search *string, limit, offset int) ([]*domain.Comment, error)
	GetChildComments(ctx context.Context, parentID int) ([]*domain.Comment, error)
}

type Service struct {
	logger         zerolog.Logger
	commentService CommentService
}

func NewService(logger zerolog.Logger, commentService CommentService) *Service {
	return &Service{
		logger:         logger,
		commentService: commentService,
	}
}
func (s *Service) Create(ctx context.Context, comment *domain.Comment) (int, error) {
	return s.commentService.Create(ctx, comment)
}
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.commentService.Delete(ctx, id)
}
func (s *Service) GetRootComments(ctx context.Context, search *string, limit, offset int) ([]*domain.Comment, error) {
	return s.commentService.GetRootComments(ctx, search, limit, offset)
}
func (s *Service) GetChildComments(ctx context.Context, parentID int) ([]*domain.Comment, error) {
	return s.commentService.GetChildComments(ctx, parentID)
}

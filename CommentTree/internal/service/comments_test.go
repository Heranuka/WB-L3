package service

import (
	"context"
	"errors"
	"testing"

	"commentTree/internal/domain"
	"commentTree/internal/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
)

func TestService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService( /* logger */ zerolog.Nop(), mockCommentService)

	comment := &domain.Comment{ID: 1, Content: "Test"}

	// Ожидаем вызова и возврат ID и nil ошибки
	mockCommentService.EXPECT().
		Create(gomock.Any(), comment).
		Return(1, nil)

	id, err := s.Create(context.Background(), comment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 1 {
		t.Errorf("expected id 1, got %d", id)
	}
}

func TestService_Create_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	comment := &domain.Comment{ID: 1, Content: "Test"}

	// Ожидаем ошибку
	mockCommentService.EXPECT().
		Create(gomock.Any(), comment).
		Return(0, errors.New("some error"))

	_, err := s.Create(context.Background(), comment)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	id := 42

	// Ожидаем вызова и возврат nil
	mockCommentService.EXPECT().
		Delete(gomock.Any(), id).
		Return(nil)

	err := s.Delete(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_Delete_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	id := 42

	// Ожидаем ошибку
	mockCommentService.EXPECT().
		Delete(gomock.Any(), id).
		Return(errors.New("delete error"))

	err := s.Delete(context.Background(), id)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_GetRootComments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	expectedComments := []*domain.Comment{
		{ID: 1, Content: "Comment 1"},
		{ID: 2, Content: "Comment 2"},
	}
	search := "test"
	limit := 10
	offset := 0

	// Ожидаем вызова и возврат данных
	mockCommentService.EXPECT().
		GetRootComments(gomock.Any(), &search, limit, offset).
		Return(expectedComments, nil)

	result, err := s.GetRootComments(context.Background(), &search, limit, offset)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 comments, got %d", len(result))
	}
}

func TestService_GetRootComments_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	search := "test"
	limit := 10
	offset := 0

	mockCommentService.EXPECT().
		GetRootComments(gomock.Any(), &search, limit, offset).
		Return(nil, errors.New("db error"))

	_, err := s.GetRootComments(context.Background(), &search, limit, offset)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_GetChildComments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	parentID := 5
	expectedComments := []*domain.Comment{
		{ID: 10, Content: "Child comment"},
	}

	mockCommentService.EXPECT().
		GetChildComments(gomock.Any(), parentID).
		Return(expectedComments, nil)

	result, err := s.GetChildComments(context.Background(), parentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 comment, got %d", len(result))
	}
}

func TestService_GetChildComments_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	parentID := 5

	mockCommentService.EXPECT().
		GetChildComments(gomock.Any(), parentID).
		Return(nil, errors.New("db error"))

	_, err := s.GetChildComments(context.Background(), parentID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_Create_NilComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	// Передача nil — ожидаем, что вызов произойдет и вернется ошибка или будет обработан корректно
	mockCommentService.EXPECT().
		Create(gomock.Any(), nil).
		Return(0, errors.New("nil comment"))

	id, err := s.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error when comment is nil, got nil")
	}
	if id != 0 {
		t.Errorf("expected id 0 for nil comment, got %d", id)
	}
}

func TestService_Delete_NonExistentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	nonExistentID := 9999

	mockCommentService.EXPECT().
		Delete(gomock.Any(), nonExistentID).
		Return(errors.New("not found"))

	err := s.Delete(context.Background(), nonExistentID)
	if err == nil || err.Error() != "not found" {
		t.Fatalf("expected 'not found' error, got %v", err)
	}
}

func TestService_Delete_UnexpectedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	id := 1
	mockCommentService.EXPECT().
		Delete(gomock.Any(), id).
		Return(errors.New("unexpected error"))

	err := s.Delete(context.Background(), id)
	if err == nil || err.Error() != "unexpected error" {
		t.Fatalf("expected 'unexpected error', got %v", err)
	}
}

func TestService_GetRootComments_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	search := "no matches"
	limit := 5
	offset := 0

	mockCommentService.EXPECT().
		GetRootComments(gomock.Any(), &search, limit, offset).
		Return([]*domain.Comment{}, nil)

	result, err := s.GetRootComments(context.Background(), &search, limit, offset)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestService_GetRootComments_InvalidParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	// Передача отрицательного лимита
	search := "test"
	limit := -1
	offset := 0

	mockCommentService.EXPECT().
		GetRootComments(gomock.Any(), &search, limit, offset).
		Return(nil, errors.New("invalid limit"))

	_, err := s.GetRootComments(context.Background(), &search, limit, offset)
	if err == nil {
		t.Fatal("expected error for invalid limit, got nil")
	}
}

func TestService_GetChildComments_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	parentID := 42

	mockCommentService.EXPECT().
		GetChildComments(gomock.Any(), parentID).
		Return([]*domain.Comment{}, nil)

	result, err := s.GetChildComments(context.Background(), parentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d", len(result))
	}
}
func TestService_Create_CancelledContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentService := mocks.NewMockCommentService(ctrl)
	s := NewService(zerolog.Nop(), mockCommentService)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // отменяем контекст перед вызовом

	comment := &domain.Comment{ID: 1, Content: "Test"}

	// В данном случае, сервис вызывает метод, но логика зависит от реализации
	mockCommentService.EXPECT().
		Create(gomock.Any(), comment).
		Return(0, errors.New("context cancelled"))

	id, err := s.Create(ctx, comment)
	if err == nil || err.Error() != "context cancelled" {
		t.Fatalf("expected 'context cancelled' error, got %v", err)
	}
	if id != 0 {
		t.Errorf("expected id 0, got %d", id)
	}
}

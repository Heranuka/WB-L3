package comments

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"commentTree/internal/domain"
	"commentTree/internal/ports/rest/comments/mocks" // путь к сгенерированным мокам
	"commentTree/pkg/e"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
)

// setupRouter создает роутер и регистрирует маршруты согласно вашему описанию
func setupRouter(t *testing.T, mockHandler CommentHandler) (*gin.Engine, func()) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	logger := zerolog.Nop()
	h := NewHandler(logger, nil, mockHandler)

	router.POST("/comments", h.CreateHandler)
	router.DELETE("/comments/:id", h.DeleteHandler)
	router.GET("/comments", h.GetRootCommentsHandler)
	router.GET("/comments/:parent_id/children", h.GetChildCommentsHandler)

	return router, func() {}
}

// Тестирование успешного создания комментария
func TestCreateHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentHandler := mocks.NewMockCommentHandler(ctrl)
	router, _ := setupRouter(t, mockCommentHandler)

	newComment := map[string]interface{}{
		"Content":  "Test comment",
		"ParentID": 0,
	}
	body, _ := json.Marshal(newComment)

	mockCommentHandler.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(1, nil)

	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["Comment Created"] != float64(1) {
		t.Errorf("unexpected response body: %v", resp)
	}
}

// Тест с ошибкой привязки JSON
func TestCreateHandler_BindError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentHandler := mocks.NewMockCommentHandler(ctrl)
	router, _ := setupRouter(t, mockCommentHandler)

	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// Успешное удаление комментария
func TestDeleteHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentHandler := mocks.NewMockCommentHandler(ctrl)
	router, _ := setupRouter(t, mockCommentHandler)

	id := 123
	mockCommentHandler.EXPECT().
		Delete(gomock.Any(), id).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/comments/"+strconv.Itoa(id), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

// Не найден комментарий для удаления
func TestDeleteHandler_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentHandler := mocks.NewMockCommentHandler(ctrl)
	router, _ := setupRouter(t, mockCommentHandler)

	id := 999
	mockCommentHandler.EXPECT().
		Delete(gomock.Any(), id).
		Return(e.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/comments/"+strconv.Itoa(id), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

// Получение корневых комментариев успешно
func TestGetRootCommentsHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentHandler := mocks.NewMockCommentHandler(ctrl)
	router, _ := setupRouter(t, mockCommentHandler)

	expectedComments := []*domain.Comment{
		{ID: 1, Content: "Comment 1"},
		{ID: 2, Content: "Comment 2"},
	}
	mockCommentHandler.EXPECT().
		GetRootComments(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(expectedComments, nil)

	req := httptest.NewRequest(http.MethodGet, "/comments?limit=2&offset=0&search=test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp []*domain.Comment
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("expected 2 comments, got %d", len(resp))
	}
}

// Получение дочерних комментариев успешно
func TestGetChildCommentsHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentHandler := mocks.NewMockCommentHandler(ctrl)
	router, _ := setupRouter(t, mockCommentHandler)

	parentID := 5
	expectedComments := []*domain.Comment{
		{ID: 10, Content: "Child 1"},
	}
	mockCommentHandler.EXPECT().
		GetChildComments(gomock.Any(), parentID).
		Return(expectedComments, nil)

	req := httptest.NewRequest(http.MethodGet, "/comments/5/children", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp []*domain.Comment
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Errorf("expected 1 comment, got %d", len(resp))
	}
}

// Некорректный ID для получения дочерних комментариев
func TestGetChildCommentsHandler_InvalidParentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentHandler := mocks.NewMockCommentHandler(ctrl)
	router, _ := setupRouter(t, mockCommentHandler)

	req := httptest.NewRequest(http.MethodGet, "/comments/abc/children", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// Ошибка при получении корневых комментариев
func TestGetRootCommentsHandler_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentHandler := mocks.NewMockCommentHandler(ctrl)
	router, _ := setupRouter(t, mockCommentHandler)

	mockCommentHandler.EXPECT().
		GetRootComments(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/comments?limit=10", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

// Ошибка при получении дочерних комментариев
func TestGetChildCommentsHandler_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommentHandler := mocks.NewMockCommentHandler(ctrl)
	router, _ := setupRouter(t, mockCommentHandler)

	mockCommentHandler.EXPECT().
		GetChildComments(gomock.Any(), 1).
		Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/comments/1/children", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

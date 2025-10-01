package items

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/wb-go/wbf/ginext"
	"wb-l3.7/internal/domain"
	mock_items "wb-l3.7/internal/ports/rest/items/mocks"
)

// setupRouterWithHandler создает роутер и регистрирует маршруты с вашим handler
func setupRouterWithHandler(handler *Handler) *ginext.Engine {
	gin.SetMode(gin.TestMode) // Используйте gin TestMode для тестов
	router := ginext.New()
	router.GET("/items", handler.GetItemsHandler)
	router.POST("/items", handler.CreateItemHandler)
	router.PUT("/items/:id", handler.UpdateItemHandler)
	router.DELETE("/items/:id", handler.DeleteItemHandler)
	router.GET("/items/history/:id", handler.GetItemHistoryHandler)
	return router
}

func TestGetItems_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemStorage := mock_items.NewMockItemStorage(ctrl)
	mockHistoryStorage := mock_items.NewMockHistoryStorage(ctrl)

	// Передаем реальный логгер, а не assert
	logger := zerolog.Nop()

	handler := NewItemsHandler(logger, mockItemStorage, mockHistoryStorage)

	now := time.Now()
	mockItemStorage.EXPECT().GetAllItems(gomock.Any()).Return([]*domain.Item{
		{
			ID: 1, Name: "Item1", Description: "Desc", Price: 10.1, Stock: 5,
			CreatedAt: now, UpdatedAt: now,
		},
	}, nil)

	router := setupRouterWithHandler(handler)

	req, _ := http.NewRequest("GET", "/items", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []domain.Item
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "Item1", resp[0].Name)
}

func TestCreateItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemStorage := mock_items.NewMockItemStorage(ctrl)
	mockHistoryStorage := mock_items.NewMockHistoryStorage(ctrl)

	logger := zerolog.Nop()
	handler := NewItemsHandler(logger, mockItemStorage, mockHistoryStorage)

	itemJSON := `{"name":"NewItem","description":"Desc","price":25.5,"stock":10}`

	mockItemStorage.EXPECT().CreateItem(gomock.Any(), gomock.Any()).Return(int64(123), nil)

	router := setupRouterWithHandler(handler)

	req, _ := http.NewRequest("POST", "/items", strings.NewReader(itemJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Item created successfully")
}

func TestUpdateItem_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemStorage := mock_items.NewMockItemStorage(ctrl)
	mockHistoryStorage := mock_items.NewMockHistoryStorage(ctrl)

	logger := zerolog.Nop()
	handler := NewItemsHandler(logger, mockItemStorage, mockHistoryStorage)

	router := setupRouterWithHandler(handler)

	req, _ := http.NewRequest("PUT", "/items/badid", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteItem_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemStorage := mock_items.NewMockItemStorage(ctrl)
	mockHistoryStorage := mock_items.NewMockHistoryStorage(ctrl)

	logger := zerolog.Nop()
	handler := NewItemsHandler(logger, mockItemStorage, mockHistoryStorage)

	mockItemStorage.EXPECT().
		DeleteItem(gomock.Any(), int64(777)).
		Return(domain.ErrNotFound)

	router := setupRouterWithHandler(handler)

	req, _ := http.NewRequest("DELETE", "/items/777", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetItemHistory_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemStorage := mock_items.NewMockItemStorage(ctrl)
	mockHistoryStorage := mock_items.NewMockHistoryStorage(ctrl)

	logger := zerolog.Nop()
	handler := NewItemsHandler(logger, mockItemStorage, mockHistoryStorage)

	historyRecords := []*domain.ItemHistoryRecord{
		{
			ID:                1,
			ItemID:            1,
			ChangedByUser:     "test", // строка, а не структура
			ChangeDescription: "Created item",
			ChangedAt:         time.Now(),
			Version:           1,
			ChangeDiff: map[string]domain.ChangeDiff{
				"price": {Old: nil, New: 100},
			},
		},
	}
	mockHistoryStorage.EXPECT().
		GetItemHistory(gomock.Any(), int64(1)).
		Return(historyRecords, nil)

	router := setupRouterWithHandler(handler)

	req, _ := http.NewRequest("GET", "/items/history/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []*domain.ItemHistoryRecord
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "Created item", resp[0].ChangeDescription)
}

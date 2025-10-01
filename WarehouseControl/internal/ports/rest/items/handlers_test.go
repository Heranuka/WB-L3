package items_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"wb-l3.7/internal/domain"
	"wb-l3.7/internal/ports/rest/items"
	mock_items "wb-l3.7/internal/ports/rest/items/mocks"

	"wb-l3.7/pkg/jwt"
)

func setupRouter(h *items.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/items/create", h.CreateItemHandler)
	r.PUT("/items/update/:id", h.UpdateItemHandler)
	r.DELETE("/items/delete/:id", h.DeleteItemHandler)
	return r
}

func addUserInfoToContext(router *gin.Engine, userID int64) {
	router.Use(func(c *gin.Context) {
		userInfo := &jwt.UserInfo{UserID: userID} // создаём userInfo с нужным ID
		ctx := context.WithValue(c.Request.Context(), "userInfo", userInfo)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
}

func TestCreateItemHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemSvc := mock_items.NewMockItemStorage(ctrl)
	mockHistSvc := mock_items.NewMockHistoryStorage(ctrl)
	handler := items.NewItemsHandler(zerolog.Nop(), mockItemSvc, mockHistSvc)
	router := setupRouter(handler)
	addUserInfoToContext(router, 42)
	testCases := []struct {
		name        string
		requestBody string
		mockSetup   func()
		wantCode    int
	}{
		{
			name:        "success",
			requestBody: `{"name":"Item1","description":"desc","price":10,"stock":5}`,
			mockSetup: func() {
				mockItemSvc.EXPECT().CreateItem(gomock.Any(), gomock.Any(), int64(42)).Return(int64(1), nil)
			},
			wantCode: http.StatusCreated,
		},
		{
			name:        "invalid json",
			requestBody: `{invalid-json}`,
			mockSetup:   func() {},
			wantCode:    http.StatusBadRequest,
		},
		{
			name:        "service error",
			requestBody: `{"name":"Item1","description":"desc","price":10,"stock":5}`,
			mockSetup: func() {
				mockItemSvc.EXPECT().CreateItem(gomock.Any(), gomock.Any(), int64(42)).Return(int64(0), errors.New("DB error"))
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/items/create", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Добавляем userInfo в контекст запроса вручную
			ctx := context.WithValue(req.Context(), "userInfo", &jwt.UserInfo{UserID: 42})
			req = req.WithContext(ctx)

			router.ServeHTTP(w, req)

			router.ServeHTTP(w, req)
			assert.Equal(t, tc.wantCode, w.Code)
		})
	}
}

func TestDeleteItemHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemSvc := mock_items.NewMockItemStorage(ctrl)
	mockHistSvc := mock_items.NewMockHistoryStorage(ctrl)
	handler := items.NewItemsHandler(zerolog.Nop(), mockItemSvc, mockHistSvc)
	router := setupRouter(handler)

	testCases := []struct {
		name      string
		paramID   string
		mockSetup func()
		wantCode  int
	}{
		{
			name:    "success",
			paramID: "5",
			mockSetup: func() {
				mockItemSvc.EXPECT().DeleteItem(gomock.Any(), int64(5)).Return(nil).Times(1)
			},
			wantCode: http.StatusOK,
		},
		{
			name:      "invalid id",
			paramID:   "bad",
			mockSetup: func() {},
			wantCode:  http.StatusBadRequest,
		},
		{
			name:    "not found",
			paramID: "6",
			mockSetup: func() {
				mockItemSvc.EXPECT().DeleteItem(gomock.Any(), int64(6)).Return(domain.ErrNotFound).Times(1)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:    "service error",
			paramID: "7",
			mockSetup: func() {
				mockItemSvc.EXPECT().DeleteItem(gomock.Any(), int64(7)).Return(errors.New("DB err")).Times(1)
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/items/delete/"+tc.paramID, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tc.wantCode, w.Code)
		})
	}
}

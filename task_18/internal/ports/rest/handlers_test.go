package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"test_18/internal/domain"
	"test_18/internal/ports/rest"
	handler_mocks "test_18/internal/ports/rest/mocks"
)

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(log.Writer(), nil))
}

func TestCreateEventHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventHandler := handler_mocks.NewMockEventHandler(ctrl)
	logger := createTestLogger()
	h := rest.NewHandler(logger, mockEventHandler)

	now := time.Now().UTC()
	reqBody := domain.Request{
		Title:       "Meeting",
		Description: "Discuss plans",
		StartTime:   domain.Date(now),
		EndTime:     domain.Date(now.Add(1 * time.Hour)),
	}

	jsonBody, _ := json.Marshal(reqBody)

	mockEventHandler.EXPECT().
		CreateEvent(gomock.Any(), gomock.Any()).
		Return(42, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/create_event", h.CreateEventHandler)

	req, _ := http.NewRequest(http.MethodPost, "/create_event", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	result := resp["result"].(map[string]interface{})
	assert.Equal(t, float64(42), result["id"])
	assert.Equal(t, "Meeting", result["title"])
}

func TestCreateEventHandler_BadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventHandler := handler_mocks.NewMockEventHandler(ctrl)
	logger := createTestLogger()
	h := rest.NewHandler(logger, mockEventHandler)

	badJson := []byte(`{"title":`)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/create_event", h.CreateEventHandler) // поменял путь на ваш

	req, _ := http.NewRequest(http.MethodPost, "/create_event", bytes.NewBuffer(badJson))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateEventHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventHandler := handler_mocks.NewMockEventHandler(ctrl)
	logger := createTestLogger()
	h := rest.NewHandler(logger, mockEventHandler)

	updateReq := domain.UpdateEventRequest{
		Title:       ptrString("New Title"),
		Description: ptrString("Updated Desc"),
	}
	jsonBody, _ := json.Marshal(updateReq)

	mockEventHandler.EXPECT().
		UpdateEvent(gomock.Any(), 123, gomock.Any()).
		Return(nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PATCH("/update_event/:id", h.UpdateEventHandler) // поменял путь и метод

	req, _ := http.NewRequest(http.MethodPatch, "/update_event/123", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(123), resp["successfully updated"])
}

func TestUpdateEventHandler_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventHandler := handler_mocks.NewMockEventHandler(ctrl)
	logger := createTestLogger()
	h := rest.NewHandler(logger, mockEventHandler)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PATCH("/update_event/:id", h.UpdateEventHandler)

	req, _ := http.NewRequest(http.MethodPatch, "/update_event/abc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteEventHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventHandler := handler_mocks.NewMockEventHandler(ctrl)
	logger := createTestLogger()
	h := rest.NewHandler(logger, mockEventHandler)

	mockEventHandler.EXPECT().
		DeleteEvent(gomock.Any(), 55).
		Return(nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/delete_event/:id", h.DeleteEventHandler) // поменял путь

	req, _ := http.NewRequest(http.MethodDelete, "/delete_event/55", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetEventsForDayHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventHandler := handler_mocks.NewMockEventHandler(ctrl)
	logger := createTestLogger()
	h := rest.NewHandler(logger, mockEventHandler)

	events := []domain.Event{
		{
			ID:        1,
			Title:     "Event 1",
			StartTime: domain.Date(time.Now().UTC()),
			EndTime:   domain.Date(time.Now().Add(time.Hour).UTC()),
		},
	}

	mockEventHandler.EXPECT().
		GetEventsForTime(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(events, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/events_for_day", h.GetEventsForDayHandler) // поменял путь

	req, _ := http.NewRequest(http.MethodGet, "/events_for_day", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp["events for day"])
}

func TestGetEventsForWeekHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventHandler := handler_mocks.NewMockEventHandler(ctrl)
	logger := createTestLogger()
	h := rest.NewHandler(logger, mockEventHandler)

	events := []domain.Event{}

	mockEventHandler.EXPECT().
		GetEventsForTime(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(events, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/events_for_week", h.GetEventsForWeekHandler)

	req, _ := http.NewRequest(http.MethodGet, "/events_for_week", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp["events for week"])
}

func TestGetEventsForMonthHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventHandler := handler_mocks.NewMockEventHandler(ctrl)
	logger := createTestLogger()
	h := rest.NewHandler(logger, mockEventHandler)

	events := []domain.Event{}

	mockEventHandler.EXPECT().
		GetEventsForTime(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(events, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/events_for_month", h.GetEventsForMonthHandler)

	req, _ := http.NewRequest(http.MethodGet, "/events_for_month", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp["events for month"])
}

// вспомогательная функция
func ptrString(s string) *string {
	return &s
}

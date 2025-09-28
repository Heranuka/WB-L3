package noteAuth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"delay/internal/domain"
	"delay/internal/ports/noteAuth"
	mock_noteAuth "delay/internal/ports/noteAuth/mocks"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// Positive CreateHandler test
func TestCreateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_noteAuth.NewMockNotificationHandler(ctrl)
	h := noteAuth.NewHandler(context.Background(), zerolog.Nop(), mockHandler)

	notification := domain.Notification{
		Message:     "Test create",
		DataToSent:  time.Now().Add(time.Hour),
		Channel:     domain.ChannelEmail,
		Destination: "test@example.com",
	}
	id := uuid.New()

	mockHandler.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(id, nil).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"message":      notification.Message,
		"data_sent_at": notification.DataToSent.Format(time.RFC3339),
		"channel":      string(notification.Channel),
		"destination":  notification.Destination,
	})
	req := httptest.NewRequest(http.MethodPost, "/notify/create", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.CreateHanlder(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp["response"])
}

// Negative CreateHandler test: invalid JSON
func TestCreateHandler_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockHandler := mock_noteAuth.NewMockNotificationHandler(ctrl)
	h := noteAuth.NewHandler(context.Background(), zerolog.Nop(), mockHandler)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/notify/create", bytes.NewReader([]byte("{invalid-json")))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.CreateHanlder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Negative CreateHandler test: missing message
func TestCreateHandler_MissingMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockHandler := mock_noteAuth.NewMockNotificationHandler(ctrl)
	h := noteAuth.NewHandler(context.Background(), zerolog.Nop(), mockHandler)

	postData := map[string]interface{}{
		"message":      "",
		"data_sent_at": time.Now().Add(time.Hour).Format(time.RFC3339),
		"channel":      string(domain.ChannelEmail),
		"destination":  "test@example.com",
	}
	reqBody, _ := json.Marshal(postData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/notify/create", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.CreateHanlder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Negative CreateHandler test: past DataToSent
func TestCreateHandler_PastDataToSent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockHandler := mock_noteAuth.NewMockNotificationHandler(ctrl)
	h := noteAuth.NewHandler(context.Background(), zerolog.Nop(), mockHandler)

	postData := map[string]interface{}{
		"message":      "test",
		"data_sent_at": time.Now().Add(-time.Hour).Format(time.RFC3339),
		"channel":      string(domain.ChannelEmail),
		"destination":  "test@example.com",
	}
	reqBody, _ := json.Marshal(postData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/notify/create", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.CreateHanlder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// GetAllHandler positive case
func TestGetAllHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_noteAuth.NewMockNotificationHandler(ctrl)
	h := noteAuth.NewHandler(context.Background(), zerolog.Nop(), mockHandler)

	notifications := []domain.Notification{
		{ID: uuid.New(), Message: "Test1", Status: "created", DataToSent: time.Now(), Channel: domain.ChannelEmail, Destination: "test1@example.com"},
	}
	mockHandler.EXPECT().
		GetAll(gomock.Any()).
		Return(&notifications, nil).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notify/all", nil)

	h.GetAllHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	notes := resp["notes"].([]interface{})
	assert.Len(t, notes, 1)
}

// GetAllHandler error case
func TestGetAllHandler_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_noteAuth.NewMockNotificationHandler(ctrl)
	h := noteAuth.NewHandler(context.Background(), zerolog.Nop(), mockHandler)

	mockHandler.EXPECT().
		GetAll(gomock.Any()).
		Return(nil, errors.New("db error")).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notify/all", nil)

	h.GetAllHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// StatusHandler positive case
func TestStatusHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_noteAuth.NewMockNotificationHandler(ctrl)
	h := noteAuth.NewHandler(context.Background(), zerolog.Nop(), mockHandler)

	id := uuid.New()
	mockHandler.EXPECT().
		Status(gomock.Any(), id).
		Return("created", nil).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notify/status/"+id.String(), nil)
	c.Params = []gin.Param{{Key: "id", Value: id.String()}}

	h.StatusHanlder(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp["response"])
}

// StatusHandler bad UUID
func TestStatusHandler_BadUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_noteAuth.NewMockNotificationHandler(ctrl)
	h := noteAuth.NewHandler(context.Background(), zerolog.Nop(), mockHandler)

	badID := "bad-uuid"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notify/status/"+badID, nil)
	c.Params = []gin.Param{{Key: "id", Value: badID}}

	h.StatusHanlder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// CancelHandler success
func TestCancelHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_noteAuth.NewMockNotificationHandler(ctrl)
	h := noteAuth.NewHandler(context.Background(), zerolog.Nop(), mockHandler)

	id := uuid.New()
	mockHandler.EXPECT().
		Cancel(gomock.Any(), id).
		Return(nil).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/notify/cancel/"+id.String(), nil)
	c.Params = []gin.Param{{Key: "id", Value: id.String()}}

	h.CancelHanlder(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// CancelHandler bad UUID
func TestCancelHandler_BadUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_noteAuth.NewMockNotificationHandler(ctrl)
	h := noteAuth.NewHandler(context.Background(), zerolog.Nop(), mockHandler)

	badID := "bad-uuid"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/notify/cancel/"+badID, nil)
	c.Params = []gin.Param{{Key: "id", Value: badID}}

	h.CancelHanlder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

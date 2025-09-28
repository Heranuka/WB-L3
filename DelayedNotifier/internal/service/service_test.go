package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"delay/internal/config"
	"delay/internal/domain"
	"delay/internal/service"
	mock_service "delay/internal/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifSvc := mock_service.NewMockNotificationService(ctrl)
	mockCache := mock_service.NewMockCacheService(ctrl)
	mockRabbit := mock_service.NewMockRabbit(ctrl)

	s := service.NewService(config.Config{}, zerolog.Nop(), mockNotifSvc, mockCache, mockRabbit)

	notification := domain.Notification{
		Message:     "Test create",
		DataToSent:  time.Now().Add(2 * time.Hour),
		Channel:     domain.ChannelEmail,
		Destination: "test@example.com",
	}
	id := uuid.New()

	mockNotifSvc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(id, nil)
	mockNotifSvc.EXPECT().UpdateStatus(gomock.Any(), id, "created").Return(nil)
	mockNotifSvc.EXPECT().Get(gomock.Any(), id).Return(&notification, nil)
	mockCache.EXPECT().Set(gomock.Any(), "notification:status:"+id.String(), "created", gomock.Any()).Return(nil)
	mockRabbit.EXPECT().Publish(&notification).Return(nil)

	gotID, err := s.Create(context.Background(), &notification)
	assert.NoError(t, err)
	assert.Equal(t, id, gotID)
}

func TestService_Create_FailCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifSvc := mock_service.NewMockNotificationService(ctrl)
	mockCache := mock_service.NewMockCacheService(ctrl)
	mockRabbit := mock_service.NewMockRabbit(ctrl)

	s := service.NewService(config.Config{}, zerolog.Nop(), mockNotifSvc, mockCache, mockRabbit)

	notification := domain.Notification{
		Message: "Fail create",
	}

	mockNotifSvc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(uuid.Nil, errors.New("create fail"))

	gotID, err := s.Create(context.Background(), &notification)
	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, gotID)
}

func TestService_Status_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifSvc := mock_service.NewMockNotificationService(ctrl)
	mockCache := mock_service.NewMockCacheService(ctrl)
	mockRabbit := mock_service.NewMockRabbit(ctrl)

	s := service.NewService(config.Config{}, zerolog.Nop(), mockNotifSvc, mockCache, mockRabbit)

	id := uuid.New()
	status := "sent"

	mockCache.EXPECT().Get(gomock.Any(), "notification:status:"+id.String(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, dest interface{}) error {
			*(dest.(*string)) = status
			return nil
		},
	)

	gotStatus, err := s.Status(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, status, gotStatus)
}

func TestService_Status_CacheMiss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifSvc := mock_service.NewMockNotificationService(ctrl)
	mockCache := mock_service.NewMockCacheService(ctrl)
	mockRabbit := mock_service.NewMockRabbit(ctrl)

	s := service.NewService(config.Config{}, zerolog.Nop(), mockNotifSvc, mockCache, mockRabbit)

	id := uuid.New()
	status := "created"

	mockCache.EXPECT().Get(gomock.Any(), "notification:status:"+id.String(), gomock.Any()).Return(errors.New("cache miss"))
	mockNotifSvc.EXPECT().Status(gomock.Any(), id).Return(status, nil)
	mockCache.EXPECT().Set(gomock.Any(), "notification:status:"+id.String(), status, gomock.Any()).Return(nil)

	gotStatus, err := s.Status(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, status, gotStatus)
}

func TestService_Cancel_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifSvc := mock_service.NewMockNotificationService(ctrl)
	mockCache := mock_service.NewMockCacheService(ctrl)
	mockRabbit := mock_service.NewMockRabbit(ctrl)

	s := service.NewService(config.Config{}, zerolog.Nop(), mockNotifSvc, mockCache, mockRabbit)

	id := uuid.New()

	mockNotifSvc.EXPECT().Cancel(gomock.Any(), id).Return(nil)
	mockCache.EXPECT().Set(gomock.Any(), "notification:status:"+id.String(), "", time.Duration(0)).Return(nil)

	err := s.Cancel(context.Background(), id)
	assert.NoError(t, err)
}

func TestService_Cancel_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifSvc := mock_service.NewMockNotificationService(ctrl)
	mockCache := mock_service.NewMockCacheService(ctrl)
	mockRabbit := mock_service.NewMockRabbit(ctrl)

	s := service.NewService(config.Config{}, zerolog.Nop(), mockNotifSvc, mockCache, mockRabbit)

	id := uuid.New()

	mockNotifSvc.EXPECT().Cancel(gomock.Any(), id).Return(errors.New("cancel error"))

	err := s.Cancel(context.Background(), id)
	assert.Error(t, err)
}

func TestService_GetAll_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifSvc := mock_service.NewMockNotificationService(ctrl)
	mockCache := mock_service.NewMockCacheService(ctrl)
	mockRabbit := mock_service.NewMockRabbit(ctrl)

	s := service.NewService(config.Config{}, zerolog.Nop(), mockNotifSvc, mockCache, mockRabbit)

	notifications := []domain.Notification{
		{ID: uuid.New(), Message: "msg1"},
		{ID: uuid.New(), Message: "msg2"},
	}

	mockCache.EXPECT().Get(gomock.Any(), "notification:getall", gomock.Any()).Return(nil).DoAndReturn(
		func(ctx context.Context, key string, dest interface{}) error {
			b, _ := json.Marshal(notifications)
			ptr := dest.(*[]domain.Notification)
			return json.Unmarshal(b, ptr)
		})

	gotNotifications, err := s.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, *gotNotifications, 2)
}

func TestService_GetAll_CacheMiss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifSvc := mock_service.NewMockNotificationService(ctrl)
	mockCache := mock_service.NewMockCacheService(ctrl)
	mockRabbit := mock_service.NewMockRabbit(ctrl)

	s := service.NewService(config.Config{}, zerolog.Nop(), mockNotifSvc, mockCache, mockRabbit)

	notifications := []domain.Notification{
		{ID: uuid.New(), Message: "msg1"},
		{ID: uuid.New(), Message: "msg2"},
	}

	mockCache.EXPECT().Get(gomock.Any(), "notification:getall", gomock.Any()).Return(errors.New("cache miss"))
	mockNotifSvc.EXPECT().GetAll(gomock.Any()).Return(&notifications, nil)
	mockCache.EXPECT().Set(gomock.Any(), "notification:getall", gomock.Any(), gomock.Any()).Return(nil)

	gotNotifications, err := s.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, *gotNotifications, 2)
}

func TestService_UpdateStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifSvc := mock_service.NewMockNotificationService(ctrl)
	mockCache := mock_service.NewMockCacheService(ctrl)
	mockRabbit := mock_service.NewMockRabbit(ctrl)

	s := service.NewService(config.Config{}, zerolog.Nop(), mockNotifSvc, mockCache, mockRabbit)

	id := uuid.New()
	status := "done"

	mockNotifSvc.EXPECT().UpdateStatus(gomock.Any(), id, status).Return(nil)

	err := s.UpdateStatus(context.Background(), id, status)
	assert.NoError(t, err)
}

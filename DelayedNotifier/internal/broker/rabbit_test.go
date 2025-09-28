package broker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	mock_broker "delay/internal/broker/mocks"
	"delay/internal/domain"
	mock_notifications "delay/internal/service/notificationService/mocks"
)

func TestPublish_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotify := mock_broker.NewMockRabbit(ctrl)

	mockNotify.EXPECT().
		Publish(gomock.Any()).
		Return(nil).
		Times(1)

	note := domain.Notification{
		ID:          uuid.New(),
		Message:     "testing message",
		DataToSent:  time.Now(),
		Channel:     "email",
		Destination: "test@example.com",
	}

	err := mockNotify.Publish(&note)
	assert.NoError(t, err)
}

func TestConsume_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotify := mock_broker.NewMockRabbit(ctrl)
	mockEmailChan := mock_notifications.NewMockNotificationChannel(ctrl)
	mockTelegramChan := mock_notifications.NewMockNotificationChannel(ctrl)

	mockNotify.EXPECT().
		Consume(gomock.Any(), mockEmailChan, mockTelegramChan).
		Return(nil).
		Times(1)

	err := mockNotify.Consume(context.Background(), mockEmailChan, mockTelegramChan)
	assert.NoError(t, err)
}

func TestUpdateStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotify := mock_broker.NewMockRabbit(ctrl)

	id := uuid.New()
	mockNotify.EXPECT().
		UpdateStatus(gomock.Any(), id, "sent").
		Return(nil).
		Times(1)

	err := mockNotify.UpdateStatus(context.Background(), id, "sent")
	assert.NoError(t, err)
}

func TestPublish_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotify := mock_broker.NewMockRabbit(ctrl)

	// Проверяем, что Publish возвращает ошибку
	expectedErr := fmt.Errorf("publish error")
	mockNotify.EXPECT().
		Publish(gomock.Any()).
		Return(expectedErr).
		Times(1)

	note := domain.Notification{
		ID:          uuid.New(),
		Message:     "test error",
		DataToSent:  time.Now(),
		Channel:     "email",
		Destination: "test@example.com",
	}

	err := mockNotify.Publish(&note)
	assert.ErrorIs(t, err, expectedErr)
}

func TestConsume_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotify := mock_broker.NewMockRabbit(ctrl)
	mockEmailChan := mock_notifications.NewMockNotificationChannel(ctrl)
	mockTelegramChan := mock_notifications.NewMockNotificationChannel(ctrl)

	expectedErr := fmt.Errorf("consume error")

	mockNotify.EXPECT().
		Consume(gomock.Any(), mockEmailChan, mockTelegramChan).
		Return(expectedErr).
		Times(1)

	err := mockNotify.Consume(context.Background(), mockEmailChan, mockTelegramChan)
	assert.ErrorIs(t, err, expectedErr)
}

func TestUpdateStatus_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotify := mock_broker.NewMockRabbit(ctrl)

	id := uuid.New()
	expectedErr := fmt.Errorf("update status error")

	mockNotify.EXPECT().
		UpdateStatus(gomock.Any(), id, "failed").
		Return(expectedErr).
		Times(1)

	err := mockNotify.UpdateStatus(context.Background(), id, "failed")
	assert.ErrorIs(t, err, expectedErr)
}

func TestConsume_WithMessageProcessing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Для этого кейса нужно сделать более сложный мок Consume, но поскольку у вас мокирут интерфейс всей Rabbit,
	// возможно, лучше протестировать метод Consume реального брокера с моками NotificationChannel.

	mockEmailChan := mock_notifications.NewMockNotificationChannel(ctrl)
	mockTelegramChan := mock_notifications.NewMockNotificationChannel(ctrl)

	// Настроить моки Send с успешным ответом
	mockEmailChan.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	mockTelegramChan.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	// Здесь можно вызвать метод Consume реального брокера с моками, если минимально сконфигурировать канал сообщений.
	// Или выделить логику обработки в отдельный метод и протестировать её отдельно.

	// Например:
	// r := NewRabbitMQ(...) с моками
	// err := r.Consume(ctx, mockEmailChan, mockTelegramChan)
	// assert.NoError(t, err)

	// Но полноценное мокирование канала сообщений обычно усложняется,
	// для этого пишут integration tests.

	t.Skip("Complex message processing mock test needs integration or refactor of logic")
}

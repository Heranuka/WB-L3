package items_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"wb-l3.7/internal/domain"
	"wb-l3.7/internal/service/items"
	"wb-l3.7/pkg/jwt"

	mock_items "wb-l3.7/internal/ports/rest/items/mocks"
)

func TestCreateItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemStorage := mock_items.NewMockItemStorage(ctrl)
	mockHistoryStorage := mock_items.NewMockHistoryStorage(ctrl)
	service := items.NewItemService(mockItemStorage, mockHistoryStorage)

	testCases := []struct {
		name          string
		userID        int64
		item          domain.Item
		mockBehavior  func()
		expectedID    int64
		expectedError error
	}{
		{
			name:   "success",
			userID: 1,
			item: domain.Item{
				Name:        "Test item",
				Description: "desc",
				Price:       10,
				Stock:       5,
			},
			mockBehavior: func() {
				mockItemStorage.EXPECT().
					CreateItem(gomock.Any(), gomock.Any(), int64(1)).
					Return(int64(100), nil)
			},
			expectedID:    100,
			expectedError: nil,
		},
		{
			name:          "missing user info",
			userID:        0,
			item:          domain.Item{Name: "No user"},
			mockBehavior:  func() {},
			expectedID:    -1,
			expectedError: fmt.Errorf("authentication info missing"),
		},
		{
			name:   "storage error",
			userID: 1,
			item: domain.Item{
				Name:        "Fail item",
				Description: "desc",
				Price:       5,
				Stock:       3,
			},
			mockBehavior: func() {
				mockItemStorage.EXPECT().
					CreateItem(gomock.Any(), gomock.Any(), int64(1)).
					Return(int64(0), errors.New("db error"))
			},
			expectedID:    -1,
			expectedError: fmt.Errorf("item service: create item failed: db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			var ctx context.Context
			if tc.userID == 0 {
				ctx = context.Background()
			} else {
				userInfo := &jwt.UserInfo{UserID: tc.userID}
				ctx = context.WithValue(context.Background(), "userInfo", userInfo)
			}

			id, err := service.CreateItem(ctx, &tc.item, tc.userID)

			assert.Equal(t, tc.expectedID, id)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestUpdateItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemStorage := mock_items.NewMockItemStorage(ctrl)
	mockHistoryStorage := mock_items.NewMockHistoryStorage(ctrl)
	service := items.NewItemService(mockItemStorage, mockHistoryStorage)

	testCases := []struct {
		name          string
		userID        int64
		item          domain.Item
		mockBehavior  func()
		expectedError error
	}{
		{
			name:   "success",
			userID: 1,
			item: domain.Item{
				ID:          100,
				Name:        "UpdatedItem",
				Description: "UpdatedDescription",
				Price:       15.0,
				Stock:       7,
			},
			mockBehavior: func() {
				mockItemStorage.EXPECT().
					UpdateItem(gomock.Any(), gomock.Any(), int64(1)).
					DoAndReturn(func(ctx context.Context, item *domain.Item, userID int64) error {
						item.UpdatedAt = time.Now()
						return nil
					}).
					Times(1) // строго один раз
			},
			expectedError: nil,
		},
		{
			name:          "missing user info",
			userID:        0,
			item:          domain.Item{ID: 101, Name: "NoUser"},
			mockBehavior:  func() {},
			expectedError: errors.New("authentication info missing"),
		},
		{
			name:   "storage error",
			userID: 1,
			item: domain.Item{
				ID:          102,
				Name:        "FailUpdate",
				Description: "Desc",
				Price:       5.0,
				Stock:       1,
			},
			mockBehavior: func() {
				mockItemStorage.EXPECT().
					UpdateItem(gomock.Any(), gomock.Any(), int64(1)).
					Return(errors.New("db error")).
					Times(1) // строго один раз
			},
			expectedError: errors.New("item service: update item failed: db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			var ctx context.Context
			if tc.userID == 0 {
				ctx = context.Background()
			} else {
				ctx = context.WithValue(context.Background(), "userInfo", &jwt.UserInfo{UserID: tc.userID})
			}

			err := service.UpdateItem(ctx, &tc.item, tc.userID)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			if err == nil {
				assert.WithinDuration(t, time.Now(), tc.item.UpdatedAt, time.Second)

			}
		})
	}

}

func TestDeleteItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemStorage := mock_items.NewMockItemStorage(ctrl)
	mockHistoryStorage := mock_items.NewMockHistoryStorage(ctrl)
	service := items.NewItemService(mockItemStorage, mockHistoryStorage)

	testCases := []struct {
		name          string
		itemID        int64
		mockBehavior  func()
		expectedError error
	}{
		{
			name:   "success",
			itemID: 100,
			mockBehavior: func() {
				mockItemStorage.EXPECT().DeleteItem(gomock.Any(), int64(100)).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "storage error",
			itemID: 101,
			mockBehavior: func() {
				mockItemStorage.EXPECT().DeleteItem(gomock.Any(), int64(101)).Return(errors.New("db error"))
			},
			expectedError: errors.New("item service: delete item failed: db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			err := service.DeleteItem(context.Background(), tc.itemID)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

package items

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"wb-l3.7/internal/domain"
	mock_items "wb-l3.7/internal/service/items/mocks"
)

func TestCreateItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemStorage := mock_items.NewMockItemStorage(ctrl)
	mockHistoryStorage := mock_items.NewMockHistoryStorage(ctrl)

	service := NewItemService(mockItemStorage, mockHistoryStorage)

	item := &domain.Item{
		Name:        "Test Product",
		Description: "Desc",
		Price:       10,
		Stock:       20,
	}

	mockItemStorage.EXPECT().
		CreateItem(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, i *domain.Item) (int64, error) {
			if i.Name != item.Name {
				t.Errorf("unexpected item name: %s", i.Name)
			}
			return 1, nil
		})

	mockHistoryStorage.EXPECT().
		LogChange(gomock.Any(), gomock.Any(), int64(1), gomock.Any(), gomock.Any()).
		Return(nil)

	ctx := context.WithValue(context.Background(), "userInfo", &domain.UserInfo{
		UserID:   1,
		Nickname: "tester",
	})

	id, err := service.CreateItem(ctx, item, userID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

func TestUpdateItem_ChangesLogged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemStorage := mock_items.NewMockItemStorage(ctrl)
	mockHistoryStorage := mock_items.NewMockHistoryStorage(ctrl)

	service := NewItemService(mockItemStorage, mockHistoryStorage)

	oldItem := &domain.Item{
		ID:          1,
		Name:        "Old Name",
		Description: "Old Desc",
		Price:       10,
		Stock:       5,
	}

	newItem := &domain.Item{
		ID:          1,
		Name:        "New Name",
		Description: "Old Desc",
		Price:       12,
		Stock:       5,
	}

	mockItemStorage.EXPECT().
		GetItem(gomock.Any(), gomock.Eq(int64(1))).
		Return(oldItem, nil)

	mockItemStorage.EXPECT().
		UpdateItem(gomock.Any(), gomock.Eq(newItem)).
		Return(nil)

	mockHistoryStorage.EXPECT().
		LogChange(gomock.Any(), gomock.Eq(int64(1)), gomock.Eq(int64(1)), gomock.Any(), gomock.Any()).
		Return(nil)

	ctx := context.WithValue(context.Background(), "userInfo", &domain.UserClaims{UserID: 1})

	err := service.UpdateItem(ctx, newItem)
	assert.NoError(t, err)
}

func TestDeleteItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemStorage := mock_items.NewMockItemStorage(ctrl)
	mockHistoryStorage := mock_items.NewMockHistoryStorage(ctrl)

	service := NewItemService(mockItemStorage, mockHistoryStorage)

	itemToDelete := &domain.Item{
		ID:   1,
		Name: "DeleteMe",
	}

	mockItemStorage.EXPECT().
		GetItem(gomock.Any(), gomock.Eq(int64(1))).
		Return(itemToDelete, nil)

	mockItemStorage.EXPECT().
		DeleteItem(gomock.Any(), gomock.Eq(int64(1))).
		Return(nil)

	mockHistoryStorage.EXPECT().
		LogChange(gomock.Any(), gomock.Eq(int64(1)), gomock.Eq(int64(1)), gomock.Any(), gomock.Nil()).
		Return(nil)

	ctx := context.WithValue(context.Background(), "userInfo", &domain.UserClaims{UserID: 1})

	err := service.DeleteItem(ctx, 1)
	assert.NoError(t, err)
}

func TestGetItemHistory_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHistoryStorage := mock_items.NewMockHistoryStorage(ctrl)
	service := NewItemService(nil, mockHistoryStorage)

	fakeHistory := []*domain.ItemHistoryRecord{
		{
			ID:                1,
			ItemID:            1,
			ChangedByUser:     "tester",
			ChangeDescription: "Created",
			ChangedAt:         time.Now(),
			Version:           1,
			ChangeDiff:        nil,
		},
	}

	mockHistoryStorage.EXPECT().
		GetItemHistory(gomock.Any(), gomock.Eq(int64(1))).
		Return(fakeHistory, nil)

	ctx := context.Background()
	hist, err := service.GetItemHistory(ctx, 1)
	assert.NoError(t, err)
	assert.Len(t, hist, 1)
}

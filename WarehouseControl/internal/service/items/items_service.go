package items

import (
	"context"
	"fmt"
	"strings"

	"wb-l3.7/internal/domain"
	"wb-l3.7/pkg/jwt"
)

//go:generate mockgen -source=items_service.go -destination=mocks/mock.go
type ItemStorage interface {
	CreateItem(ctx context.Context, item *domain.Item, userID int64) (int64, error)
	GetItem(ctx context.Context, itemID int64) (*domain.Item, error)
	GetAllItems(ctx context.Context) ([]*domain.Item, error)
	UpdateItem(ctx context.Context, item *domain.Item, userID int64) error
	DeleteItem(ctx context.Context, itemID int64) error
}

type HistoryStorage interface {
	LogChange(ctx context.Context, userID, itemID int64, changeDesc string, changeDiff map[string]domain.ChangeDiff) error
	GetItemHistory(ctx context.Context, itemID int64) ([]*domain.ItemHistoryRecord, error)
}

type ItemService struct {
	itemStorage    ItemStorage
	historyStorage HistoryStorage
}

func NewItemService(itemStorage ItemStorage, historyStorage HistoryStorage) *ItemService {
	return &ItemService{
		itemStorage:    itemStorage,
		historyStorage: historyStorage,
	}
}
func (s *ItemService) CreateItem(ctx context.Context, item *domain.Item, userID int64) (int64, error) {
	userInfo, exists := jwt.GetUserInfoFromContext(ctx)
	if !exists {
		return -1, fmt.Errorf("authentication info missing")
	}

	if userInfo.UserID != userID {
		return -1, fmt.Errorf("userID mismatch: user from context %d vs param %d", userInfo.UserID, userID)
	}
	itemID, err := s.itemStorage.CreateItem(ctx, item, userID)
	if err != nil {
		return -1, fmt.Errorf("item service: create item failed: %w", err)
	}

	/*
	   // Создаем changeDiff для создания
	   changeDiff := map[string]domain.ChangeDiff{
	       "name":        {Old: nil, New: item.Name},
	       "description": {Old: nil, New: item.Description},
	       "price":       {Old: nil, New: item.Price},
	       "stock":       {Old: nil, New: item.Stock},
	   }

	   err = s.historyStorage.LogChange(ctx, userInfo.UserID, itemID, fmt.Sprintf("Item created: %s", item.Name), changeDiff)
	   if err != nil {
	       fmt.Printf("Failed to log item creation for item %d: %v\n", itemID, err)
	   }
	*/

	return itemID, nil
}
func (s *ItemService) UpdateItem(ctx context.Context, item *domain.Item, userID int64) error {
	userInfo, exists := jwt.GetUserInfoFromContext(ctx)
	if !exists {
		return fmt.Errorf("authentication info missing")
	}
	if userInfo.UserID != userID {
		return fmt.Errorf("userID mismatch: user from context %d vs param %d", userInfo.UserID, userID)
	}

	err := s.itemStorage.UpdateItem(ctx, item, userID)
	if err != nil {
		return fmt.Errorf("item service: update item failed: %w", err)
	}

	return nil
}

func (s *ItemService) DeleteItem(ctx context.Context, itemID int64) error {

	/* itemToDelete, err := s.itemStorage.GetItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("item service: get item for delete failed: %w", err)
	} */

	err := s.itemStorage.DeleteItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("item service: delete item failed: %w", err)
	}

	/*
	   err = s.historyStorage.LogChange(ctx, userInfo.UserID, itemID, fmt.Sprintf("Item deleted: %s", itemToDelete.Name), nil)
	   if err != nil {
	       fmt.Printf("Failed to log item deletion for item %d: %v\n", itemID, err)
	   }
	*/

	return nil
}

func (s *ItemService) GenerateChangeDescriptionAndDiff(oldItem, newItem *domain.Item) (string, map[string]domain.ChangeDiff) {
	var changes []string
	changeDiff := make(map[string]domain.ChangeDiff)

	if oldItem.Name != newItem.Name {
		changes = append(changes, fmt.Sprintf("name changed from '%s' to '%s'", oldItem.Name, newItem.Name))
		changeDiff["name"] = domain.ChangeDiff{Old: oldItem.Name, New: newItem.Name}
	}
	if oldItem.Description != newItem.Description {
		changes = append(changes, fmt.Sprintf("description changed from '%s' to '%s'", oldItem.Description, newItem.Description))
		changeDiff["description"] = domain.ChangeDiff{Old: oldItem.Description, New: newItem.Description}
	}
	if oldItem.Price != newItem.Price {
		changes = append(changes, fmt.Sprintf("price changed from %.2f to %.2f", oldItem.Price, newItem.Price))
		changeDiff["price"] = domain.ChangeDiff{Old: oldItem.Price, New: newItem.Price}
	}
	if oldItem.Stock != newItem.Stock {
		changes = append(changes, fmt.Sprintf("stock changed from %d to %d", oldItem.Stock, newItem.Stock))
		changeDiff["stock"] = domain.ChangeDiff{Old: oldItem.Stock, New: newItem.Stock}
	}

	return strings.Join(changes, "; "), changeDiff
}

func (s *ItemService) GetItemHistory(ctx context.Context, itemID int64) ([]*domain.ItemHistoryRecord, error) {
	history, err := s.historyStorage.GetItemHistory(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("item service: get item history failed: %w", err)
	}
	return history, nil
}

func (s *ItemService) LogChange(ctx context.Context, userID, itemID int64, changeDesc string, changeDiff map[string]domain.ChangeDiff) error {
	return s.historyStorage.LogChange(ctx, userID, itemID, changeDesc, changeDiff)
}

func (s *ItemService) GetAllItems(ctx context.Context) ([]*domain.Item, error) {
	return s.itemStorage.GetAllItems(ctx)
}

func (s *ItemService) GetItem(ctx context.Context, itemID int64) (*domain.Item, error) {
	return s.itemStorage.GetItem(ctx, itemID)
}

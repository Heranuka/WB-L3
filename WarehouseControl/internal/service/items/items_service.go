package items

import (
	"context"
	"fmt"
	"strings"
	"time"

	"wb-l3.7/internal/domain"
	"wb-l3.7/pkg/jwt"
)

type ItemStorage interface {
	CreateItem(ctx context.Context, item *domain.Item) (int64, error)
	GetItem(ctx context.Context, itemID int64) (*domain.Item, error)
	GetAllItems(ctx context.Context) ([]*domain.Item, error)
	UpdateItem(ctx context.Context, item *domain.Item) error
	DeleteItem(ctx context.Context, itemID int64) error
}

type HistoryStorage interface {
	LogChange(ctx context.Context, userID, itemID int64, changeDescription string) error
	GetItemHistory(ctx context.Context, itemID int64) ([]*domain.ItemHistoryEntry, error)
}

type ItemService struct {
	itemStorage    ItemStorage
	historyStorage HistoryStorage
}

func (s *ItemService) CreateItem(ctx context.Context, item *domain.Item) (int64, error) {

	userInfo, _ := jwt.GetUserInfoFromContext(ctx)
	if userInfo == nil {
		return -1, fmt.Errorf("authentication info missing")
	}

	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	itemID, err := s.itemStorage.CreateItem(ctx, item)
	if err != nil {
		return -1, fmt.Errorf("item service: create item failed: %w", err)
	}

	err = s.historyStorage.LogChange(ctx, userInfo.UserID, itemID, fmt.Sprintf("Item created: %s", item.Name))
	if err != nil {
		fmt.Printf("Failed to log item creation for item %d: %v\n", itemID, err)
	}

	return itemID, nil
}

func NewItemService(itemStorage ItemStorage, historyStorage HistoryStorage) *ItemService {
	return &ItemService{
		itemStorage:    itemStorage,
		historyStorage: historyStorage,
	}
}

func (s *ItemService) GetItem(ctx context.Context, itemID int64) (*domain.Item, error) {
	item, err := s.itemStorage.GetItem(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("item service: get item failed: %w", err)
	}
	return item, nil
}

func (s *ItemService) GetAllItems(ctx context.Context) ([]*domain.Item, error) {
	items, err := s.itemStorage.GetAllItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("item service: get all items failed: %w", err)
	}
	return items, nil
}

func (s *ItemService) UpdateItem(ctx context.Context, item *domain.Item) error {
	userInfo, _ := jwt.GetUserInfoFromContext(ctx)
	if userInfo == nil {
		return fmt.Errorf("authentication info missing")
	}

	currentItem, err := s.itemStorage.GetItem(ctx, item.ID)
	if err != nil {
		return fmt.Errorf("item service: get item for update failed: %w", err)
	}

	item.UpdatedAt = time.Now()
	err = s.itemStorage.UpdateItem(ctx, item)
	if err != nil {
		return fmt.Errorf("item service: update item failed: %w", err)
	}

	changes := s.generateChangeDescription(currentItem, item)
	if changes != "" {
		err = s.historyStorage.LogChange(ctx, userInfo.UserID, item.ID, changes)
		if err != nil {
			fmt.Printf("Failed to log item update for item %d: %v\n", item.ID, err)
		}
	}

	return nil
}

func (s *ItemService) DeleteItem(ctx context.Context, itemID int64) error {
	userInfo, _ := jwt.GetUserInfoFromContext(ctx)
	if userInfo == nil {
		return fmt.Errorf("authentication info missing")
	}

	itemToDelete, err := s.itemStorage.GetItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("item service: get item for delete failed: %w", err)
	}

	err = s.itemStorage.DeleteItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("item service: delete item failed: %w", err)
	}

	err = s.historyStorage.LogChange(ctx, userInfo.UserID, itemID, fmt.Sprintf("Item deleted: %s", itemToDelete.Name))
	if err != nil {
		fmt.Printf("Failed to log item deletion for item %d: %v\n", itemID, err)
	}

	return nil
}

func (s *ItemService) generateChangeDescription(oldItem, newItem *domain.Item) string {
	var changes []string
	if oldItem.Name != newItem.Name {
		changes = append(changes, fmt.Sprintf("name changed from '%s' to '%s'", oldItem.Name, newItem.Name))
	}
	if oldItem.Description != newItem.Description {
		changes = append(changes, fmt.Sprintf("description changed from '%s' to '%s'", oldItem.Description, newItem.Description))
	}
	if oldItem.Price != newItem.Price {
		changes = append(changes, fmt.Sprintf("price changed from %.2f to %.2f", oldItem.Price, newItem.Price))
	}
	if oldItem.Stock != newItem.Stock {
		changes = append(changes, fmt.Sprintf("stock changed from %d to %d", oldItem.Stock, newItem.Stock))
	}

	return strings.Join(changes, "; ")
}

func (s *ItemService) GetItemHistory(ctx context.Context, itemID int64) ([]*domain.ItemHistoryEntry, error) {
	history, err := s.historyStorage.GetItemHistory(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("item service: get item history failed: %w", err)
	}
	return history, nil
}

func (s *ItemService) LogChange(ctx context.Context, userId, itemID int64, changeDescription string) error {
	return s.historyStorage.LogChange(ctx, userId, itemID, changeDescription)
}

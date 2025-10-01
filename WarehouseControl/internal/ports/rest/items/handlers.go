package items

import (
	"context"
	"net/http"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/wb-go/wbf/ginext"
	"wb-l3.7/internal/domain"
	"wb-l3.7/pkg/jwt"
)

//go:generate mockgen -source=handlers.go -destination=mocks/mock.go
type ItemStorage interface {
	CreateItem(ctx context.Context, item *domain.Item, userID int64) (int64, error)
	GetItem(ctx context.Context, itemID int64) (*domain.Item, error)
	GetAllItems(ctx context.Context) ([]*domain.Item, error)
	UpdateItem(ctx context.Context, item *domain.Item, userID int64) error
	DeleteItem(ctx context.Context, itemID int64, userID int64) error
}

type HistoryStorage interface {
	LogChange(ctx context.Context, userID, itemID int64, changeDesc string, changeDiff map[string]domain.ChangeDiff) error
	GetItemHistory(ctx context.Context, itemID int64) ([]*domain.ItemHistoryRecord, error)
}

// Handler - Gin handlers for items and their history
type Handler struct {
	logger         zerolog.Logger
	itemService    ItemStorage
	historyStorage HistoryStorage
}

// NewItemsHandler creates new handler instance
func NewItemsHandler(
	logger zerolog.Logger,
	itemService ItemStorage,
	historyStorage HistoryStorage,
) *Handler {
	return &Handler{
		logger:         logger,
		itemService:    itemService,
		historyStorage: historyStorage,
	}
}

func (h *Handler) GetUserProfileHandler(c *ginext.Context) {
	userInfo, exists := c.Get("userInfo")
	if !exists {
		h.logger.Error().Msg("authentication info not found in context")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "authentication info not found"})
		return
	}
	userClaims, ok := userInfo.(*jwt.UserInfo)
	if !ok {
		h.logger.Error().Msg("invalid user info type in context")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "invalid user info type"})
		return
	}

	h.logger.Info().Int64("userID", userClaims.UserID).Msg("User profile requested")

	c.JSON(http.StatusOK, ginext.H{
		"id":       userClaims.UserID,
		"nickname": userClaims.Nickname,
		"roles":    userClaims.Roles,
	})
}

func (h *Handler) CreateItemHandler(c *ginext.Context) {
	var req domain.Item
	userInfo, exists := c.Get("userInfo")
	if !exists {
		h.logger.Warn().Msg("user_id is not correct")
		c.JSON(http.StatusBadRequest, "user_id is not correct")
		return
	}
	userID := userInfo.(*jwt.UserInfo).UserID

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn().Err(err).Msg("Invalid create item request")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	itemID, err := h.itemService.CreateItem(c.Request.Context(), &req, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create item")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to create item"})
		return
	}

	h.logger.Info().Int64("itemID", itemID).Msg("Item created successfully")

	c.JSON(http.StatusCreated, ginext.H{"message": "Item created successfully", "id": itemID})
}

func (h *Handler) GetItemsHandler(c *ginext.Context) {
	items, err := h.itemService.GetAllItems(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to retrieve items")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to get items"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetItemHandler(c *ginext.Context) {
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		h.logger.Warn().Err(err).Str("param", itemIDStr).Msg("Invalid item ID parameter")
		c.JSON(http.StatusNotFound, ginext.H{"error": "invalid item ID"})
		return
	}

	item, err := h.itemService.GetItem(c.Request.Context(), itemID)
	if err != nil {
		if err == domain.ErrNotFound {
			h.logger.Info().Int64("itemID", itemID).Msg("Item not found")
			c.JSON(http.StatusNotFound, ginext.H{"error": "item not found"})
			return
		}
		h.logger.Error().Err(err).Int64("itemID", itemID).Msg("Failed to get item")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to get item"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateItemHandler(c *ginext.Context) {
	userInfo, exists := c.Get("userInfo")
	if !exists {
		h.logger.Warn().Msg("user_id is not correct")
		c.JSON(http.StatusBadRequest, "user_id is not correct")
		return
	}
	userID := userInfo.(*jwt.UserInfo).UserID
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		h.logger.Warn().Err(err).Str("param", itemIDStr).Msg("Invalid item ID parameter")
		c.JSON(http.StatusNotFound, ginext.H{"error": "invalid item ID"})
		return
	}

	var req domain.Item
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn().Err(err).Msg("Invalid update request body")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if req.ID != 0 && req.ID != itemID {
		h.logger.Warn().
			Int64("bodyID", req.ID).
			Int64("paramID", itemID).
			Msg("Item ID mismatch between param and body")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "item ID mismatch"})
		return
	}
	req.ID = itemID

	err = h.itemService.UpdateItem(c.Request.Context(), &req, userID)
	if err != nil {
		if err == domain.ErrNotFound {
			h.logger.Info().Int64("itemID", itemID).Msg("Item to update not found")
			c.JSON(http.StatusNotFound, ginext.H{"error": "item not found"})
			return
		}
		h.logger.Error().Err(err).Int64("itemID", itemID).Msg("Failed to update item")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to update item"})
		return
	}

	h.logger.Info().Int64("itemID", itemID).Msg("Item updated successfully")
	c.JSON(http.StatusOK, ginext.H{"message": "Item updated successfully"})
}

func (h *Handler) DeleteItemHandler(c *ginext.Context) {
	userInfo, exists := c.Get("userInfo")
	if !exists {
		h.logger.Warn().Msg("user_id is not correct")
		c.JSON(http.StatusBadRequest, "user_id is not correct")
		return
	}
	userID := userInfo.(*jwt.UserInfo).UserID
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		h.logger.Warn().Err(err).Str("param", itemIDStr).Msg("Invalid item ID parameter")
		c.JSON(http.StatusNotFound, ginext.H{"error": "invalid item ID"})
		return
	}

	err = h.itemService.DeleteItem(c.Request.Context(), itemID, userID)
	if err != nil {
		if err == domain.ErrNotFound {
			h.logger.Info().Int64("itemID", itemID).Msg("Item to delete not found")
			c.JSON(http.StatusNotFound, ginext.H{"error": "item not found"})
			return
		}
		h.logger.Error().Err(err).Int64("itemID", itemID).Msg("Failed to delete item")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to delete item"})
		return
	}

	h.logger.Info().Int64("itemID", itemID).Msg("Item deleted successfully")
	c.JSON(http.StatusOK, ginext.H{"message": "Item deleted successfully"})
}

func (h *Handler) GetItemHistoryHandler(c *ginext.Context) {
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		h.logger.Warn().Err(err).Str("param", itemIDStr).Msg("Invalid item ID for history request")
		c.JSON(http.StatusNotFound, ginext.H{"error": "invalid item ID"})
		return
	}

	history, err := h.historyStorage.GetItemHistory(c.Request.Context(), itemID)
	if err != nil {
		if err == domain.ErrNotFound {
			h.logger.Info().Int64("itemID", itemID).Msg("Item history not found")
			c.JSON(http.StatusNotFound, ginext.H{"error": "item not found"})
			return
		}
		h.logger.Error().Err(err).Int64("itemID", itemID).Msg("Failed to get item history")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to get item history"})
		return
	}

	h.logger.Info().Int64("itemID", itemID).Msg("Item history retrieved successfully")
	c.JSON(http.StatusOK, history)
}

/*
func (h *Handler) GetLogChangeHandler(c *ginext.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn().Err(err).Str("param", userIDStr).Msg("Invalid user ID for log change")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid user ID"})
		return
	}

	// Пример, placeholder для changeDescription и changeDiff:
	changeDescription := "Sample change description"
	changeDiff := map[string]domain.ChangeDiff{
		"field1": {Old: "oldValue", New: "newValue"},
	}

	err = h.historyStorage.LogChange(c.Request.Context(), userID, 2, changeDescription, changeDiff)
	if err != nil {
		if err == domain.ErrNotFound {
			h.logger.Info().Int64("userID", userID).Msg("User not found for log change")
			c.JSON(http.StatusNotFound, ginext.H{"error": "user not found"})
			return
		}
		h.logger.Error().Err(err).Int64("userID", userID).Msg("Failed to log change")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to log change"})
		return
	}

	h.logger.Info().Int64("userID", userID).Msg("Change logged successfully")
	c.JSON(http.StatusOK, ginext.H{"success": "success"})
}
*/

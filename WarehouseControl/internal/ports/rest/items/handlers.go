package items

import (
	"context"
	"errors"
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
	DeleteItem(ctx context.Context, itemID int64) error
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

// GetUserProfileHandler godoc
// @Summary Получение профиля пользователя
// @Description Возвращает информацию о текущем пользователе по токену
// @Tags items
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Информация о пользователе"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /profile [get]
func (h *Handler) GetUserProfileHandler(c *ginext.Context) {
	userInfo, exists := jwt.GetUserInfoFromContext(c.Request.Context())
	if !exists {
		h.logger.Error().Msg("authentication info not found in context")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "authentication info not found"})
		return
	}

	h.logger.Info().Int64("userID", userInfo.UserID).Msg("User profile requested")

	c.JSON(http.StatusOK, ginext.H{
		"id":       userInfo.UserID,
		"nickname": userInfo.Nickname,
		"roles":    userInfo.Roles,
	})
}

// @Summary Create a new item
// @Description Creates a new item with the provided details.
// @Tags Items
// @Accept json
// @Produce json
// @Param item body domain.Item true "Item data to create"
// @Param Authorization header string true "Bearer JWT Token"
// @Security BearerAuth
// @Success 201 {object} ginext.H{"message": string, "id": integer} "Item created successfully"
// @Failure 400 {object} ginext.H{"error": string} "Invalid request body or missing user ID"
// @Failure 401 {object} ginext.H{"error": string} "Unauthorized - Missing or invalid authentication token"
// @Failure 403 {object} ginext.H{"error": string} "Forbidden - User does not have the required permissions"
// @Failure 500 {object} ginext.H{"error": string} "Internal server error"
// @Router /items/create [post]
func (h *Handler) CreateItemHandler(c *ginext.Context) {
	var req domain.Item
	userInfo, exists := jwt.GetUserInfoFromContext(c.Request.Context())
	if !exists {
		h.logger.Warn().Msg("user_id is not correct")
		c.JSON(http.StatusBadRequest, "user_id is not correct")
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn().Err(err).Msg("Invalid create item request")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	itemID, err := h.itemService.CreateItem(c.Request.Context(), &req, userInfo.UserID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create item")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to create item"})
		return
	}

	h.logger.Info().Int64("itemID", itemID).Msg("Item created successfully")

	c.JSON(http.StatusCreated, ginext.H{"message": "Item created successfully", "id": itemID})
}

// @Summary Get all items
// @Description Retrieves a list of all items available in the system.
// @Tags Items
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.Item "A list of items"
// @Failure 401 {object} ginext.H{"error": string} "Unauthorized - Missing or invalid authentication token"
// @Failure 403 {object} ginext.H{"error": string} "Forbidden - User does not have the required permissions"
// @Failure 500 {object} ginext.H{"error": string} "Internal server error"
// @Router /items/getall [get]
func (h *Handler) GetItemsHandler(c *ginext.Context) {
	items, err := h.itemService.GetAllItems(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to retrieve items")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to get items"})
		return
	}
	c.JSON(http.StatusOK, items)
}

/*
// GetItemHandler godoc
// @Summary Получить предмет по ID
// @Description Возвращает предмет по его ID
// @Tags items
// @Accept json
// @Produce json
// @Param id path int true "ID предмета"
// @Success 200 {object} domain.Item
// @Failure 404 {object} map[string]string "Предмет не найден"
// @Failure 500 {object} map[string]string "Ошибка сервиса"
// @Router /items/{id} [get]
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
} */

// @Summary Update an item by ID
// @Description Updates an existing item in the system.
// @Tags Items
// @Accept json
// @Produce json
// @Param id path int true "The unique identifier of the item to update"
// @Param item body domain.Item true "The item data to update"
// @Security BearerAuth
// @Success 200 {object} ginext.H{"message": string} "Item updated successfully"
// @Failure 400 {object} ginext.H{"error": string} "Invalid item ID supplied or invalid request body"
// @Failure 401 {object} ginext.H{"error": string} "Unauthorized - Missing or invalid authentication token"
// @Failure 403 {object} ginext.H{"error": string} "Forbidden - User does not have the required permissions"
// @Failure 404 {object} ginext.H{"error": string} "Item not found"
// @Failure 500 {object} ginext.H{"error": string} "Internal server error"
// @Router /items/update/{id} [put]

func (h *Handler) UpdateItemHandler(c *ginext.Context) {
	userInfo, exists := jwt.GetUserInfoFromContext(c.Request.Context())
	if !exists {
		h.logger.Warn().Msg("authentication info missing")
		c.JSON(http.StatusBadRequest, "authentication info missing")
		return
	}

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

	err = h.itemService.UpdateItem(c.Request.Context(), &req, userInfo.UserID)
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

// @Summary Delete an item by ID
// @Description Deletes a specific item from the system by its unique identifier.
// @Tags Items
// @Accept json
// @Produce json
// @Param id path int true "The unique identifier of the item to delete"
// @Security BearerAuth
// @Success 200 {object} object{message=string} "Item deleted successfully"
// @Failure 400 {object} ginext.H{"error": string} "Invalid item ID supplied"
// @Failure 401 {object} ginext.H{"error": string} "Unauthorized - Missing or invalid authentication token"
// @Failure 403 {object} ginext.H{"error": string} "Forbidden - User does not have the required permissions"
// @Failure 404 {object} ginext.H{"error": string} "Item not found"
// @Failure 500 {object} ginext.H{"error": string} "Internal server error"
// @Router /items/delete/{id} [delete]
func (h *Handler) DeleteItemHandler(c *ginext.Context) {
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		h.logger.Warn().Err(err).Str("param", itemIDStr).Msg("Invalid item ID parameter")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid item ID"})
		return
	}

	err = h.itemService.DeleteItem(c.Request.Context(), itemID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
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

// GetItemHistoryHandler godoc
// @Summary История изменений предмета
// @Description Возвращает историю изменений выбранного предмета
// @Tags items
// @Accept json
// @Produce json
// @Param id path int true "ID предмета"
// @Success 200 {array} domain.ItemHistoryRecord
// @Failure 404 {object} map[string]string "История не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервиса"
// @Router /items/history/{id} [get]
func (h *Handler) GetItemHistoryHandler(c *ginext.Context) {
	itemIDStr := c.Param("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		h.logger.Warn().Err(err).Str("param", itemIDStr).Msg("Invalid item ID for history request")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid item ID"})
		return
	}

	history, err := h.historyStorage.GetItemHistory(c.Request.Context(), itemID)
	if err != nil || len(history) <= 0 {
		if err == domain.ErrNotFound {
			h.logger.Info().Msg("Item history not found")
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

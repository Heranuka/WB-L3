package items

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	// Для работы с ID из URL
	"github.com/gin-gonic/gin"
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

	// Дополнительно (для поиска/фильтрации)
}

// Handler - Gin хэндлеры для аутентификации и авторизации
type Handler struct {
	logger         *slog.Logger
	itemService    ItemStorage
	historyStorage HistoryStorage // Сервис для работы с товарами (CRUD, история)

}

// NewAuthHandler создает новый экземпляр хэндлера.
// Предполагается, что Auth Handler будет использовать Auth, Item, User и JWT сервисы.
func NewItemsHandler(
	logger *slog.Logger,
	itemService ItemStorage, historyStorage HistoryStorage,
) *Handler {
	return &Handler{
		logger:         logger,
		itemService:    itemService,
		historyStorage: historyStorage,
	}
}

// GetUserProfileHandler godoc
// @Summary Получить профиль пользователя
// @Description Возвращает информацию о текущем пользователе из JWT
// @Tags profile
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /profile [get]
func (h *Handler) GetUserProfileHandler(c *gin.Context) {
	// Получаем UserInfo из контекста, которое было добавлено ValidateTokenMiddleware
	userInfo, exists := c.Get("userInfo")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication info not found"})
		return
	}
	userClaims, ok := userInfo.(*jwt.UserInfo)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user info type"})
		return
	}

	// Можно вернуть полную информацию о пользователе, если она доступна в UserInfo
	// Или, если нужна более детальная информация, вызвать UserService
	c.JSON(http.StatusOK, gin.H{
		"id":       userClaims.UserID,
		"nickname": userClaims.Nickname,
		"roles":    userClaims.Roles,
	})
}

// CreateItemHandler godoc
// @Summary Создать новый товар
// @Description Создает новый товар и возвращает его id
// @Tags items
// @Accept json
// @Produce json
// @Param item body domain.Item true "Данные товара"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /items/create [post]
func (h *Handler) CreateItemHandler(c *gin.Context) {
	var req domain.Item // Предполагаем, что приходит только name, description, price, stock
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	itemID, err := h.itemService.CreateItem(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create item", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create item"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Item created successfully", "id": itemID})
}

// GetItemsHandler godoc
// @Summary Получить список товаров
// @Description Возвращает список всех товаров
// @Tags items
// @Produce json
// @Success 200 {array} domain.Item
// @Failure 500 {object} map[string]string
// @Router /items/getll [get]
func (h *Handler) GetItemsHandler(c *gin.Context) {
	items, err := h.itemService.GetAllItems(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get items", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get items"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetItemHandler godoc
// @Summary Получить товар по ID
// @Description Возвращает товар по заданному ID
// @Tags items
// @Produce json
// @Param id path int true "ID товара"
// @Success 200 {object} domain.Item
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /items/{id} [get]
func (h *Handler) GetItemHandler(c *gin.Context) { // Для GET /items/{id}
	itemIDStr := c.Param("id") // Получаем ID товара из URL
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	item, err := h.itemService.GetItem(c.Request.Context(), itemID)
	if err != nil {
		if err == domain.ErrNotFound { // Предполагаем, что такая ошибка есть
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}
		h.logger.Error("Failed to get item", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get item"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// UpdateItemHandler godoc
// @Summary Обновить товар
// @Description Обновляет товар по ID
// @Tags items
// @Accept json
// @Produce json
// @Param id path int true "ID товара"
// @Param item body domain.Item true "Данные для обновления"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /items/update/{id} [put]
func (h *Handler) UpdateItemHandler(c *gin.Context) {
	itemIDStr := c.Param("id") // Получаем ID товара из URL
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	var req domain.Item // Принимаем полную структуру для обновления
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Убеждаемся, что ID из URL совпадает с ID из тела запроса, если он есть
	if req.ID != 0 && req.ID != itemID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item ID mismatch"})
		return
	}
	req.ID = itemID // Устанавливаем ID из URL

	err = h.itemService.UpdateItem(c.Request.Context(), &req)
	if err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}
		h.logger.Error("Failed to update item", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item updated successfully"})
}

// DeleteItemHandler godoc
// @Summary Удалить товар
// @Description Удаляет товар по ID
// @Tags items
// @Produce json
// @Param id path int true "ID товара"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /items/delete/{id} [delete]
func (h *Handler) DeleteItemHandler(c *gin.Context) {
	itemIDStr := c.Param("id") // Получаем ID товара из URL
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	err = h.itemService.DeleteItem(c.Request.Context(), itemID)
	if err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}
		h.logger.Error("Failed to delete item", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully"})
}

// GetItemHistoryHandler godoc
// @Summary Получить историю изменений товара
// @Description Возвращает историю изменений товара по ID
// @Tags items history
// @Produce json
// @Param id path int true "ID товара"
// @Success 200 {array} domain.ItemHistoryEntry
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /items/history/{id} [get]
func (h *Handler) GetItemHistoryHandler(c *gin.Context) {
	itemIDStr := c.Param("id") // Получаем ID товара из URL
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	history, err := h.historyStorage.GetItemHistory(c.Request.Context(), itemID)
	if err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}
		h.logger.Error("Failed to get item history", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get item history"})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetLogChangeHandler godoc
// @Summary Добавить запись в лог изменений
// @Description Логирует изменение для пользователя по user_id
// @Tags history
// @Produce json
// @Param user_id path int true "ID пользователя"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /logchange/{user_id} [get]
func (h *Handler) GetLogChangeHandler(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Invalid user id in token", slog.String("error", err.Error()))
	}

	err = h.historyStorage.LogChange(c.Request.Context(), userID, 2, "2")
	if err == domain.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "success"})
}

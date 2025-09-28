package noteAuth

import (
	"context"
	"delay/internal/domain"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/wb-go/wbf/ginext"
)

//go:generate mockgen -source=handlers.go -destination=mocks/mock.go
type NotificationHandler interface {
	Create(ctx context.Context, notification *domain.Notification) (uuid.UUID, error)
	Status(ctx context.Context, noteID uuid.UUID) (string, error)
	Cancel(ctx context.Context, noteID uuid.UUID) error
	GetAll(ctx context.Context) (*[]domain.Notification, error)
}

type Handler struct {
	logger        zerolog.Logger
	notifyHandler NotificationHandler
}

func NewHandler(ctx context.Context, logger zerolog.Logger, actions NotificationHandler) *Handler {
	return &Handler{
		logger:        logger,
		notifyHandler: actions,
	}
}

// GetAllHandler godoc
// @Summary Get all notifications
// @Description Get a list of all notifications
// @Tags notifications
// @Produce  json
// @Success 200 {object} map[string][]domain.Notification "List of notifications"
// @Failure 400 {object} map[string]string "Bad request"
// @Router /notify/all [get]
func (h *Handler) GetAllHandler(c *ginext.Context) {
	notes, err := h.notifyHandler.GetAll(c.Request.Context())
	if err != nil {
		h.logger.Error().
			Err(err).
			Msg("Failed GetAll")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "failed to GetAll"})
		return
	}
	c.JSON(http.StatusOK, ginext.H{"notes": notes})
}

// CreateHanlder godoc
// @Summary Create a new notification
// @Description Create a notification with message, channel, destination, and delivery time
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body RequestNote true "Notification request body"
// @Success 200 {object} map[string]domain.Notification "Created notification"
// @Failure 400 {object} map[string]string "Bad request, validation errors"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notify/create [post]
func (h *Handler) CreateHanlder(c *ginext.Context) {
	var request RequestNote
	if err := c.Bind(&request); err != nil {
		h.logger.Error().
			Err(err).
			Msg("Failed to decode the request struct")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Could not decode the data into request"})
		return
	}

	if request.Message == "" {
		h.logger.Error().
			Msg("Message is required")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Message is required"})
		return
	}

	if request.DataToSent.Before(time.Now()) {
		h.logger.Warn().
			Msg("DataToSent must be in the future")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "DataToSent must be in the future"})
		return
	}

	notification := domain.Notification{
		Message:     request.Message,
		DataToSent:  request.DataToSent,
		Channel:     request.Channel,
		Destination: request.Destination,
	}
	noteID, err := h.notifyHandler.Create(c.Request.Context(), &notification)
	if err != nil {
		h.logger.Error().
			Err(err).
			Msg("Error creating notification")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Failed to create notification"})
		return
	}

	response := domain.Notification{
		ID:          noteID,
		Message:     request.Message,
		DataToSent:  request.DataToSent,
		Channel:     request.Channel,
		Destination: request.Destination,
		CreatedAt:   time.Now(),
	}
	c.JSON(http.StatusOK, ginext.H{"response": response})
}

// StatusHanlder godoc
// @Summary Get notification status by ID
// @Description Retrieve status of a notification by its UUID
// @Tags notifications
// @Produce json
// @Param id path string true "Notification UUID"
// @Success 200 {object} map[string]domain.StatusResponse "Status response"
// @Failure 400 {object} map[string]string "Bad request or not found"
// @Router /notify/status/{id} [get]
func (h *Handler) StatusHanlder(c *ginext.Context) {
	str := c.Param("id")
	id, err := uuid.Parse(str)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("idParam", str).
			Msg("Error parsing id")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid id format"})
		return
	}

	status, err := h.notifyHandler.Status(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "noteID not found" {
			c.JSON(http.StatusBadRequest, ginext.H{"error": "id not found"})
			return
		}
		h.logger.Error().
			Err(err).
			Str("noteID", id.String()).
			Msg("Error getting status")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "failed to get status"})
		return
	}

	response := domain.StatusResponse{
		NoteID: id,
		Status: status,
	}
	c.JSON(http.StatusOK, ginext.H{"response": response})
}

// CancelHanlder godoc
// @Summary Cancel (delete) a notification by ID
// @Description Cancel notification identified by UUID
// @Tags notifications
// @Produce json
// @Param id path string true "Notification UUID"
// @Success 200 {object} map[string]domain.CancelResponse "Cancel confirmation"
// @Failure 400 {object} map[string]string "Bad request or not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notify/cancel/{id} [delete]
func (h *Handler) CancelHanlder(c *ginext.Context) {
	str := c.Param("id")
	id, err := uuid.Parse(str)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("idParam", str).
			Msg("Ошибка парсинга UUID")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid UUID"})
		return
	}
	if err := h.notifyHandler.Cancel(c.Request.Context(), id); err != nil {
		if err.Error() == "noteID not found" {
			c.JSON(http.StatusBadRequest, ginext.H{"error": "id not found to delete"})
			return
		}
		h.logger.Error().
			Err(err).
			Str("id", id.String()).
			Msg("Error deleting notification")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to delete notification"})
		return
	}
	response := domain.CancelResponse{
		NoteID:  id,
		Message: "The notification successfully deleted",
	}
	c.JSON(http.StatusOK, ginext.H{"response": response})
}

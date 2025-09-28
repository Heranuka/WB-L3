package rest

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"test_18/internal/domain"
	"time"

	"github.com/gin-gonic/gin"
)

//go:generate mockgen -source=handlers.go -destination=mocks/mock.go
type EventHandler interface {
	CreateEvent(ctx context.Context, event domain.Event) (int, error)
	UpdateEvent(ctx context.Context, id int, req domain.UpdateEventRequest) error
	DeleteEvent(ctx context.Context, id int) error
	GetEventsForDay(ctx context.Context, userID int, date time.Time) (domain.User, error)
	GetEventsForWeek(ctx context.Context, userID int, date time.Time) (domain.User, error)
	GetEventsForMonth(ctx context.Context, userID int, date time.Time) (domain.User, error)
	CreateUser(ctx context.Context) (int, error)
}

type Handler struct {
	logger       *slog.Logger
	eventHandler EventHandler
}

func NewHandler(logger *slog.Logger, eventHandler EventHandler) *Handler {
	return &Handler{
		logger:       logger,
		eventHandler: eventHandler,
	}
}

func (h *Handler) CreateEventHandler(c *gin.Context) {
	var req domain.Request

	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind in CreateEvent", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.eventHandler.CreateUser(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to create user", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	event := domain.Event{
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
		EventDate:   domain.Date(time.Time(req.EventDate).In(time.UTC)),
		CreatedAt:   domain.Date(time.Now().UTC()),
		UpdatedAt:   domain.Date(time.Now().UTC()),
	}

	eventID, err := h.eventHandler.CreateEvent(c.Request.Context(), event)
	if err != nil {
		h.logger.Error("Failed to CreateEvent", slog.String("error", err.Error()), slog.Int("user_id", userID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event.ID = eventID

	c.JSON(http.StatusOK, gin.H{"result": event})
}

func (h *Handler) UpdateEventHandler(c *gin.Context) {
	var req domain.UpdateEventRequest

	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		h.logger.Error("invalid id", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind UpdateEventRequest", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.eventHandler.UpdateEvent(c.Request.Context(), id, req); err != nil {
		h.logger.Error("Failed to UpdateEvent", slog.String("error", err.Error()), slog.Int("id", id))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"successfully updated": id})
}

func (h *Handler) DeleteEventHandler(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		h.logger.Error("invalid id", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.eventHandler.DeleteEvent(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to DeleteEvent", slog.String("error", err.Error()), slog.Int("id", id))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"successfully deleted": id})
}

func parseUserIDAndDate(c *gin.Context) (int, time.Time, error) {
	dateParam := c.Query("date")
	if dateParam == "" {
		dateParam = time.Now().UTC().Format("2006-01-02")
	}

	idstr := c.Query("user_id")
	if idstr == "" {
		return 0, time.Time{}, fmt.Errorf("missing user_id parameter")
	}
	id, err := strconv.Atoi(idstr)
	if err != nil {
		return 0, time.Time{}, err
	}

	date, err := time.Parse("2006-01-02", dateParam)
	if err != nil {
		return 0, time.Time{}, err
	}
	date = date.UTC()

	return id, date, nil
}
func (h *Handler) GetEventsForDayHandler(c *gin.Context) {
	userID, date, err := parseUserIDAndDate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.eventHandler.GetEventsForDay(c.Request.Context(), userID, date)
	if err != nil {
		h.logger.Error("Failed to GetEventsForDay", slog.String("error", err.Error()))
		statusCode := http.StatusInternalServerError
		if err.Error() == "not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("GetEventsForDay success", slog.Any("user", user))
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) GetEventsForWeekHandler(c *gin.Context) {
	userID, date, err := parseUserIDAndDate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.eventHandler.GetEventsForWeek(c.Request.Context(), userID, date)
	if err != nil {
		h.logger.Error("Failed to GetEventsForWeek", slog.String("error", err.Error()))
		statusCode := http.StatusInternalServerError
		if err.Error() == "not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("GetEventsForWeek success", slog.Any("user", user))
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) GetEventsForMonthHandler(c *gin.Context) {
	userID, date, err := parseUserIDAndDate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.eventHandler.GetEventsForMonth(c.Request.Context(), userID, date)
	if err != nil {
		h.logger.Error("Failed to GetEventsForMonth", slog.String("error", err.Error()))
		statusCode := http.StatusInternalServerError
		if err.Error() == "not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("GetEventsForMonth success", slog.Any("user", user))
	c.JSON(http.StatusOK, gin.H{"user": user})
}

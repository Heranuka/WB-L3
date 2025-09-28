package rest

import (
	"commentTree/internal/domain"
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentHandler interface {
	Create(ctx context.Context, comment *domain.Comment) (int, error)
	GetById(ctx context.Context, id int) (*domain.Comment, error)
	Delete(ctx context.Context, id int) error
}

type Handler struct {
	logger         *slog.Logger
	commentHandler CommentHandler
}

func NewHandler(logger *slog.Logger, commentHandler CommentHandler) *Handler {
	return &Handler{
		logger:         logger,
		commentHandler: commentHandler,
	}
}

func (h *Handler) CreateHandler(c *gin.Context) {
	var comment *domain.Comment

	if err := c.Bind(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.commentHandler.Create(c, comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Comment Created": id})
}

func (h *Handler) GetByIdHandler(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comment, err := h.commentHandler.GetById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comment)
}

func (h *Handler) DeleteHandler(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.commentHandler.Delete(c, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Successfully deleted": id})
}

package comments

import (
	"commentTree/internal/config"
	"commentTree/internal/domain"
	"commentTree/pkg/e"
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/wb-go/wbf/ginext"
)

//go:generate mockgen -source=handlers.go -destination=mocks/mock.go
type CommentHandler interface {
	Create(ctx context.Context, comment *domain.Comment) (int, error)
	Delete(ctx context.Context, id int) error
	GetRootComments(ctx context.Context, search *string, limit, offset int) ([]*domain.Comment, error)
	GetChildComments(ctx context.Context, parentID int) ([]*domain.Comment, error)
}

type Handler struct {
	logger         zerolog.Logger
	cfg            *config.Config
	commentHandler CommentHandler
}

func NewHandler(logger zerolog.Logger, config *config.Config, commentHandler CommentHandler) *Handler {
	return &Handler{
		logger:         logger,
		cfg:            config,
		commentHandler: commentHandler,
	}
}
func (h *Handler) CreateHandler(c *ginext.Context) {
	var newComment commentCreate

	if err := c.ShouldBindJSON(&newComment); err != nil {
		h.logger.Error().Err(err).Msg("CreateHandler: failed to bind JSON")
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	comment := &domain.Comment{
		Content:  newComment.Content,
		ParentID: newComment.ParentID,
	}

	id, err := h.commentHandler.Create(c.Request.Context(), comment)
	if err != nil {
		h.logger.Error().Err(err).Msg("CreateHandler: failed to create comment")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ginext.H{"Comment Created": id})
}

func (h *Handler) DeleteHandler(c *ginext.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		h.logger.Error().Err(err).Msg("DeleteHandler: invalid id param")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid id"})
		return
	}
	if err := h.commentHandler.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, e.ErrNotFound) {
			h.logger.Warn().Msgf("DeleteHandler: comment id %d not found", id)
			c.JSON(http.StatusNotFound, e.ErrNotFound)
			return
		}
		h.logger.Error().Err(err).Msgf("DeleteHandler: failed to delete comment id %d", id)
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, "deleted successfully")
}

func (h *Handler) GetRootCommentsHandler(c *ginext.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	search := c.Query("search")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	comments, err := h.commentHandler.GetRootComments(c.Request.Context(), &search, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("GetRootCommentsHandler: failed to get root comments")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comments)
}

func (h *Handler) GetChildCommentsHandler(c *ginext.Context) {
	parentIDStr := c.Param("parent_id")
	parentID, err := strconv.Atoi(parentIDStr)
	if err != nil {
		if errors.Is(err, e.ErrNotFound) {
			h.logger.Warn().Msgf("GetChildCommentsHandler: invalid parent ID %s not found", parentIDStr)
			c.JSON(http.StatusNotFound, e.ErrNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("GetChildCommentsHandler: invalid parent ID")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid parent ID"})
		return
	}

	comments, err := h.commentHandler.GetChildComments(c.Request.Context(), parentID)
	if err != nil {
		h.logger.Error().Err(err).Msgf("GetChildCommentsHandler: failed to get children for parent ID %d", parentID)
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comments)
}

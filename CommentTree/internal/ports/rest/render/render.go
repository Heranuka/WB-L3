package render_handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HandlerRender interface {
	Home(w http.ResponseWriter)
}

type Handler struct {
	logger        *slog.Logger
	handlerRender HandlerRender
}

func NewHandler(logger *slog.Logger, handlerRender HandlerRender) *Handler {
	return &Handler{
		logger:        logger,
		handlerRender: handlerRender,
	}
}

func (h *Handler) HomeHandler(c *gin.Context) {
	h.handlerRender.Home(c.Writer)
}

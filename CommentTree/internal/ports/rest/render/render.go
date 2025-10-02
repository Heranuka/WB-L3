package render_handler

import (
	"net/http"

	"github.com/rs/zerolog"
	"github.com/wb-go/wbf/ginext"
)

type HandlerRender interface {
	Home(w http.ResponseWriter)
}

type Handler struct {
	logger        zerolog.Logger
	handlerRender HandlerRender
}

func NewHandler(logger zerolog.Logger, handlerRender HandlerRender) *Handler {
	return &Handler{
		logger:        logger,
		handlerRender: handlerRender,
	}
}

func (h *Handler) HomeHandler(c *ginext.Context) {
	h.handlerRender.Home(c.Writer)
}

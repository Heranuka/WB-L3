package render

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

type RenderHandler interface {
	Home(w http.ResponseWriter)
}
type Handler struct {
	render RenderHandler
}

func NewHandler(render RenderHandler) *Handler {
	return &Handler{
		render: render,
	}
}

func (h *Handler) Home(c *ginext.Context) {
	h.render.Home(c.Writer)
}

package renderer

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
	"wb-l3.7/pkg/jwt"
)

type RenderHandler interface {
	Home(w http.ResponseWriter, data any)
	LoginPage(w http.ResponseWriter)
	RegisterPage(w http.ResponseWriter)
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
	user, exists := c.Get("userInfo")
	if !exists {
		h.render.Home(c.Writer, ginext.H{"User": nil})
		return
	}
	userInfo, ok := user.(*jwt.UserInfo)
	if !ok {
		h.render.Home(c.Writer, ginext.H{"User": nil})
		return
	}
	h.render.Home(c.Writer, ginext.H{"User": userInfo})
}

func (h *Handler) Loginpage(c *ginext.Context) {
	h.render.LoginPage(c.Writer)
}

func (h *Handler) Registerpage(c *ginext.Context) {
	h.render.RegisterPage(c.Writer)
}

func (a *Handler) RootRedirect(c *ginext.Context) {
	userInfo, exists := c.Get("userInfo")
	if !exists || userInfo == nil {
		// user not authorized — redirect to /register
		c.Redirect(http.StatusFound, "/register")
		return
	}
	// user authorized — redirect to home
	c.Redirect(http.StatusFound, "/home")
}

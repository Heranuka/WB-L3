package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"wb-l3.7/internal/domain"
	"wb-l3.7/internal/service/render"
	"wb-l3.7/pkg/jwt"
)

type ServiceAuth interface {
	Register(ctx context.Context, nickname, password string, roles []jwt.Role) error
	Login(ctx context.Context, nickname, password string) (*domain.Tokens, *domain.User, error)
	Refresh(ctx context.Context, token string) (*domain.Tokens, error)
}

type Handler struct {
	logger *slog.Logger
	auth   ServiceAuth
	render render.Render
}

func NewHandler(logger *slog.Logger, authService ServiceAuth, render render.Render) *Handler {
	return &Handler{
		logger: logger,
		auth:   authService,
		render: render,
	}
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Регистрация нового пользователя с ролью
// @Tags auth
// @Accept json
// @Produce json
// @Param register body registerRequest true "Данные для регистрации"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /user/register [post]
func (h *Handler) Register(c *gin.Context) {
	var register registerRequest

	if err := c.Bind(&register); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roleToAssign := register.Roles

	if len(roleToAssign) == 0 {
		roleToAssign = []jwt.Role{jwt.Viewer}
	}

	if err := validator.New().Struct(register); err != nil {
		var validateErrs validator.ValidationErrors
		errors.As(err, &validateErrs)

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.auth.Register(c.Request.Context(), register.Nickname, register.Password, roleToAssign)
	if err != nil {
		if errors.Is(err, domain.ErrNicknameAlreadyExist) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("failed to register user", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Success": "success"})
}

// Login godoc
// @Summary Авторизация пользователя
// @Description Авторизация и получение JWT токенов
// @Tags auth
// @Accept json
// @Produce json
// @Param login body loginRequest true "Данные для логина"
// @Success 200 {object} tokenResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/login [post]
func (h *Handler) Login(c *gin.Context) {
	var logReq loginRequest

	if err := c.Bind(&logReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validator.New().Struct(logReq); err != nil {
		var validateErrs validator.ValidationErrors
		errors.As(err, &validateErrs)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, user, err := h.auth.Login(c.Request.Context(), logReq.Nickname, logReq.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("failed to login user", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tokenResp := tokenResponse{
		UserID:       user.ID,
		Nickname:     user.Nickname,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		CreatedAt:    tokens.CreatedAt,
		ExpiresAt:    tokens.ExpiresAt,
		Roles:        user.Roles,
	}

	c.SetCookie(
		"jwt_token",
		tokens.AccessToken,
		3600*24, // время жизни cookie в секундах (например, сутки)
		"/",
		"localhost", // домен
		false,       // secure (true если на https)
		true,        // httpOnly - недоступен JS
	)
	c.JSON(http.StatusOK, tokenResp)
}

// RefreshToken godoc
// @Summary Обновление JWT токенов
// @Description Обновить access и refresh токен с помощью refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body refreshRequest true "Refresh token для обновления"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /user/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var refresh refreshRequest

	if err := c.Bind(&refresh); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validator.New().Struct(refresh); err != nil {
		var validateErrs validator.ValidationErrors
		errors.As(err, &validateErrs)

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.auth.Refresh(c.Request.Context(), refresh.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("failed to refresh tokens", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenResp := tokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}
	c.JSON(http.StatusOK, gin.H{"success": tokenResp})
}

func (h *Handler) Homepage(c *gin.Context) {
	user, exists := c.Get("userInfo")
	if !exists {
		c.HTML(http.StatusOK, "home.html", gin.H{"User": nil})
		return
	}

	if user == nil {
		c.HTML(http.StatusOK, "home.html", gin.H{"User": nil})
		return
	}

	c.HTML(http.StatusOK, "home.html", gin.H{"User": user})
}

func (h *Handler) Loginpage(c *gin.Context) {
	h.render.LoginPage(c.Writer)
}

func (h *Handler) Registerpage(c *gin.Context) {
	h.render.RegisterPage(c.Writer)
}

func (a *Handler) RootRedirect(c *gin.Context) {
	userInfo, exists := c.Get("userInfo")
	if !exists || userInfo == nil {
		// Пользователь не авторизован — редирект на /login
		c.Redirect(http.StatusFound, "/register")
		return
	}
	// Пользователь авторизован — редирект на домашнюю страницу
	c.Redirect(http.StatusFound, "/home")
}

package linksAuth

import (
	"context"
	"fmt"
	"net/http"
	"shortener/internal/domain"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

//go:generate mockgen -source=handlers.go -destination=mocks/mock.go
type ShortLinkService interface {
	Create(ctx context.Context, link *domain.ShortURL) error
	Get(ctx context.Context, shortURL string) (domain.ShortURL, error)
}

type ClickService interface {
	Save(ctx context.Context, click *domain.Click) error
	AggregateByDay(ctx context.Context, shortURL string) ([]domain.DayStats, error)
	AggregateByMonth(ctx context.Context, shortURL string) ([]domain.MonthStats, error)
	AggregateByUserAgent(ctx context.Context, shortURL string) ([]domain.UserAgentStats, error)
}

type Handler struct {
	logger           zerolog.Logger
	shortLinkService ShortLinkService
	clickService     ClickService
}

func NewHandler(logger zerolog.Logger, shortLinkService ShortLinkService,
	clickService ClickService) *Handler {
	return &Handler{
		logger:           logger,
		shortLinkService: shortLinkService,
		clickService:     clickService,
	}
}

// CreateShortLinkRequest запрос на создание короткой ссылки
type CreateShortLinkRequest struct {
	URL       string `json:"url" binding:"required,url" example:"https://example.com"`
	ShortCode string `json:"short_code,omitempty" example:"customcode"`
}

// Create
// @Summary Создать короткую ссылку
// @Description Создает короткую ссылку с необязательным кастомным кодом
// @Tags ShortLink
// @Accept json
// @Produce json
// @Param request body CreateShortLinkRequest true "Request body"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /shorten [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().
			Err(err).
			Msg("invalid create request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	shortLink := &domain.ShortURL{
		OriginalURL: req.URL,
		CreatedAt:   time.Now(),
	}
	if req.ShortCode != "" {
		shortLink.ShortCode = req.ShortCode
		shortLink.Custom = true
	}

	if err := h.shortLinkService.Create(c.Request.Context(), shortLink); err != nil {
		h.logger.Error().
			Err(err).
			Str("url", req.URL).
			Str("short_code", req.ShortCode).
			Msg("failed to create short link")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	} else if forwarded := c.Request.Header.Get("X-Forwarded-Proto"); forwarded != "" {
		scheme = forwarded
	}
	host := c.Request.Host
	fullShortURL := fmt.Sprintf("%s://%s/s/%s", scheme, host, shortLink.ShortCode)

	h.logger.Info().
		Str("url", req.URL).
		Str("short_url", shortLink.ShortCode).
		Msg("short link created successfully")

	c.JSON(http.StatusOK, gin.H{
		"short_url":    fullShortURL,
		"original_url": shortLink.OriginalURL,
		"created_at":   shortLink.CreatedAt,
	})
}

// Redirect
// @Summary Redirect to original URL
// @Description Redirects user by short URL code
// @Tags ShortLink
// @Param shortURL path string true "Short URL code"
// @Success 302 {string} string "Redirect"
// @Failure 404 {object} map[string]string
// @Router /s/{shortURL} [get]
func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("shortURL")

	link, err := h.shortLinkService.Get(c.Request.Context(), code)
	if err != nil {
		h.logger.Warn().
			Str("short_url", code).
			Msg("short URL not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	// Сохраняем информацию о переходе
	click := &domain.Click{
		ShortURLID: link.ID,
		Timestamp:  time.Now(),
		UserAgent:  c.Request.UserAgent(),
		IPAddress:  c.ClientIP(),
	}
	if err := h.clickService.Save(c.Request.Context(), click); err != nil {
		h.logger.Error().
			Err(err).
			Str("short_url", code).
			Msg("failed to save click info")
		// не прерываем редирект из-за ошибки логирования клика
	}

	h.logger.Info().
		Str("short_url", code).
		Str("redirect_to", link.OriginalURL).
		Msg("redirecting to original URL")

	c.Redirect(http.StatusFound, link.OriginalURL)
}

// Analytics
// @Summary Get analytics for short URL
// @Description Returns stats grouped by day, month, user-agent
// @Tags ShortLink
// @Param shortURL path string true "Short URL code"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /analytics/{shortURL} [get]
func (h *Handler) Analytics(c *gin.Context) {
	shortURL := c.Param("shortURL")

	h.logger.Info().Str("short_url", shortURL).Msg("analytics request")

	statsDay, err := h.clickService.AggregateByDay(c.Request.Context(), shortURL)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("short_url", shortURL).
			Msg("failed to get daily stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get daily stats"})
		return
	}

	statsMonth, err := h.clickService.AggregateByMonth(c.Request.Context(), shortURL)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("short_url", shortURL).
			Msg("failed to get monthly stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get monthly stats"})
		return
	}

	statsUA, err := h.clickService.AggregateByUserAgent(c.Request.Context(), shortURL)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("short_url", shortURL).
			Msg("failed to get user agent stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user agent stats"})
		return
	}

	h.logger.Info().
		Str("short_url", shortURL).
		Msg("analytics retrieved successfully")

	c.JSON(http.StatusOK, gin.H{
		"short_url":          shortURL,
		"stats_by_day":       statsDay,
		"stats_by_month":     statsMonth,
		"stats_by_useragent": statsUA,
	})
}

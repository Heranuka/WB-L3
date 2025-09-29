package linksAuth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"shortener/internal/domain"
	mock_linksAuth "shortener/internal/ports/rest/linksAuth/mocks"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func setupRouter(t *testing.T) (*gin.Engine, *mock_linksAuth.MockShortLinkService, *mock_linksAuth.MockClickService, *Handler) {
	ctrl := gomock.NewController(t)
	shortSvc := mock_linksAuth.NewMockShortLinkService(ctrl)
	clickSvc := mock_linksAuth.NewMockClickService(ctrl)
	logger := zerolog.Nop()
	handler := NewHandler(logger, shortSvc, clickSvc)

	r := gin.Default()
	r.POST("/shorten", handler.Create)
	r.GET("/s/:shortURL", handler.Redirect)
	r.GET("/analytics/:shortURL", handler.Analytics)

	return r, shortSvc, clickSvc, handler
}

func TestCreate_Success(t *testing.T) {
	r, shortSvc, _, _ := setupRouter(t)

	shortSvc.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil).
		DoAndReturn(func(ctx context.Context, link *domain.ShortURL) error {
			link.ShortCode = "abc123"
			return nil
		})

	reqJSON := `{"url": "https://example.com", "short_code": "abc123"}`
	req, _ := http.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"short_url"`)
	assert.Contains(t, w.Body.String(), `"original_url":"https://example.com"`)
}

func TestCreate_InvalidJSON(t *testing.T) {
	r, _, _, _ := setupRouter(t)

	reqJSON := `{"url": "not a url"}`
	req, _ := http.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error"`)
}

func TestRedirect_Success(t *testing.T) {
	r, shortSvc, clickSvc, _ := setupRouter(t)

	shortSvc.EXPECT().
		Get(gomock.Any(), "abc123").
		Return(domain.ShortURL{ID: 1, OriginalURL: "https://example.com", ShortCode: "abc123"}, nil)

	clickSvc.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(nil)

	req, _ := http.NewRequest(http.MethodGet, "/s/abc123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://example.com", w.Result().Header.Get("Location"))
}

func TestRedirect_NotFound(t *testing.T) {
	r, shortSvc, _, _ := setupRouter(t)

	shortSvc.EXPECT().
		Get(gomock.Any(), "unknown").
		Return(domain.ShortURL{}, errors.New("not found"))

	req, _ := http.NewRequest(http.MethodGet, "/s/unknown", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `Short URL not found`)
}

func TestAnalytics_Success(t *testing.T) {
	r, _, clickSvc, _ := setupRouter(t)

	dayStats := []domain.DayStats{
		{Date: time.Now(), Count: 5},
	}
	monthStats := []domain.MonthStats{
		{Year: 2025, Month: 9, Count: 20},
	}
	uaStats := []domain.UserAgentStats{
		{UserAgent: "Mozilla", Count: 10},
	}

	clickSvc.EXPECT().AggregateByDay(gomock.Any(), "abc123").Return(dayStats, nil)
	clickSvc.EXPECT().AggregateByMonth(gomock.Any(), "abc123").Return(monthStats, nil)
	clickSvc.EXPECT().AggregateByUserAgent(gomock.Any(), "abc123").Return(uaStats, nil)

	req, _ := http.NewRequest(http.MethodGet, "/analytics/abc123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"stats_by_day"`)
	assert.Contains(t, w.Body.String(), `"stats_by_month"`)
	assert.Contains(t, w.Body.String(), `"stats_by_useragent"`)
}

func TestAnalytics_AggregateByDayError(t *testing.T) {
	r, _, clickSvc, _ := setupRouter(t)

	clickSvc.EXPECT().AggregateByDay(gomock.Any(), "abc123").Return(nil, errors.New("db error"))

	req, _ := http.NewRequest(http.MethodGet, "/analytics/abc123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `Failed to get daily stats`)
}

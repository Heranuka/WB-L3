package url_shortener

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"shortener/internal/domain"
	"time"

	"github.com/rs/zerolog"
)

//go:generate mockgen -source=url_shortener.go -destination=mocks/mock.go
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
}

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

type URLShortener struct {
	logger           zerolog.Logger
	shortLinkService ShortLinkService
	clickService     ClickService
	cache            Cache
}

func NewURLShortener(logger zerolog.Logger, shortLinkService ShortLinkService,
	clickService ClickService, cache Cache) *URLShortener {
	return &URLShortener{
		logger:           logger,
		shortLinkService: shortLinkService,
		clickService:     clickService,
		cache:            cache,
	}
}

func (s *URLShortener) Create(ctx context.Context, link *domain.ShortURL) error {
	cachedLink := &domain.ShortURL{}
	if err := s.cache.Get(ctx, link.OriginalURL, cachedLink); err == nil {
		*link = *cachedLink
		return nil
	}

	existingLink, err := s.shortLinkService.Get(ctx, link.OriginalURL)
	if err == nil {
		*link = existingLink
		_ = s.cache.Set(ctx, link.OriginalURL, link, time.Hour*24)
		return nil
	}
	if err != nil && !errors.Is(err, domain.ErrLinkNotFound) {
		return err
	}

	if link.Custom && link.ShortCode != "" {
		_, err := s.shortLinkService.Get(ctx, link.ShortCode)
		if err == nil {
			return fmt.Errorf("custom short code %s already exists", link.ShortCode)
		}
		if err != nil && !errors.Is(err, domain.ErrLinkNotFound) {
			return err
		}
	} else {
		link.ShortCode = generateShortURL(6)
	}

	if err := s.shortLinkService.Create(ctx, link); err != nil {
		return err
	}

	_ = s.cache.Set(ctx, link.OriginalURL, link, time.Hour*24)
	_ = s.cache.Set(ctx, link.ShortCode, link, time.Hour*24)
	return nil
}

func (s *URLShortener) Get(ctx context.Context, shortURL string) (domain.ShortURL, error) {
	var link domain.ShortURL
	if err := s.cache.Get(ctx, shortURL, &link); err == nil {
		return link, nil
	}

	link, err := s.shortLinkService.Get(ctx, shortURL)
	if err != nil {
		return domain.ShortURL{}, err
	}

	_ = s.cache.Set(ctx, shortURL, link, time.Hour*24)
	return link, nil
}

func (s *URLShortener) AggregateByDay(ctx context.Context, shortURL string) ([]domain.DayStats, error) {
	return s.clickService.AggregateByDay(ctx, shortURL)
}

func (s *URLShortener) AggregateByMonth(ctx context.Context, shortURL string) ([]domain.MonthStats, error) {
	return s.clickService.AggregateByMonth(ctx, shortURL)
}

func (s *URLShortener) AggregateByUserAgent(ctx context.Context, shortURL string) ([]domain.UserAgentStats, error) {
	return s.clickService.AggregateByUserAgent(ctx, shortURL)
}

func (s *URLShortener) Save(ctx context.Context, click *domain.Click) error {
	return s.clickService.Save(ctx, click)
}

func generateShortURL(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

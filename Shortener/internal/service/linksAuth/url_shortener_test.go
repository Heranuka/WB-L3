package url_shortener

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"shortener/internal/domain"
	mock_url_shortener "shortener/internal/service/linksAuth/mocks"
)

func TestCreate_NewLinkSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortSvc := mock_url_shortener.NewMockShortLinkService(ctrl)
	mockClickSvc := mock_url_shortener.NewMockClickService(ctrl)
	mockCache := mock_url_shortener.NewMockCache(ctrl)

	u := NewURLShortener(zerolog.Nop(), mockShortSvc, mockClickSvc, mockCache)

	ctx := context.Background()
	originalURL := "https://example.com"
	link := &domain.ShortURL{
		OriginalURL: originalURL,
	}

	// Мокируем кэш (нет значения)
	mockCache.EXPECT().Get(ctx, originalURL, gomock.Any()).Return(errors.New("cache miss"))
	mockShortSvc.EXPECT().Get(ctx, originalURL).Return(domain.ShortURL{}, domain.ErrLinkNotFound)
	mockShortSvc.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, link *domain.ShortURL) error {
		// Принудительно задан shortCode для теста
		link.ShortCode = "abc123"
		return nil
	})
	mockCache.EXPECT().Set(gomock.Any(), gomock.Eq(originalURL), gomock.Any(), gomock.Any()).Return(nil)
	mockCache.EXPECT().Set(gomock.Any(), gomock.Eq("abc123"), gomock.Any(), gomock.Any()).Return(nil)

	err := u.Create(ctx, link)
	assert.NoError(t, err)
	assert.NotEmpty(t, link.ShortCode)
}

func TestCreate_CustomCodeExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortSvc := mock_url_shortener.NewMockShortLinkService(ctrl)
	mockClickSvc := mock_url_shortener.NewMockClickService(ctrl)
	mockCache := mock_url_shortener.NewMockCache(ctrl)

	u := NewURLShortener(zerolog.Nop(), mockShortSvc, mockClickSvc, mockCache)

	ctx := context.Background()
	link := &domain.ShortURL{
		OriginalURL: "https://example.com",
		ShortCode:   "abc123",
		Custom:      true,
	}

	mockCache.EXPECT().Get(ctx, link.OriginalURL, gomock.Any()).Return(errors.New("cache miss"))

	mockShortSvc.EXPECT().Get(ctx, link.OriginalURL).Return(domain.ShortURL{}, domain.ErrLinkNotFound)

	// Проверяем, что кастомный shortCode уже существует
	mockShortSvc.EXPECT().Get(ctx, "abc123").Return(domain.ShortURL{ShortCode: "abc123"}, nil)

	err := u.Create(ctx, link)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "custom short code abc123 already exists")
}

func TestGetCacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortSvc := mock_url_shortener.NewMockShortLinkService(ctrl)
	mockClickSvc := mock_url_shortener.NewMockClickService(ctrl)
	mockCache := mock_url_shortener.NewMockCache(ctrl)

	u := NewURLShortener(zerolog.Nop(), mockShortSvc, mockClickSvc, mockCache)

	ctx := context.Background()
	expectedLink := domain.ShortURL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
	}

	mockCache.EXPECT().Get(ctx, "abc123", gomock.Any()).DoAndReturn(func(_ context.Context, _ string, dest interface{}) error {
		p, ok := dest.(*domain.ShortURL)
		if !ok {
			return errors.New("wrong type")
		}
		*p = expectedLink
		return nil
	})

	link, err := u.Get(ctx, "abc123")
	assert.NoError(t, err)
	assert.Equal(t, expectedLink, link)
}

func TestGetCacheMiss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortSvc := mock_url_shortener.NewMockShortLinkService(ctrl)
	mockClickSvc := mock_url_shortener.NewMockClickService(ctrl)
	mockCache := mock_url_shortener.NewMockCache(ctrl)

	u := NewURLShortener(zerolog.Nop(), mockShortSvc, mockClickSvc, mockCache)

	ctx := context.Background()
	expectedLink := domain.ShortURL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
	}

	mockCache.EXPECT().Get(ctx, "abc123", gomock.Any()).Return(errors.New("cache miss"))

	mockShortSvc.EXPECT().Get(ctx, "abc123").Return(expectedLink, nil)

	mockCache.EXPECT().Set(ctx, "abc123", expectedLink, gomock.Any()).Return(nil)

	link, err := u.Get(ctx, "abc123")
	assert.NoError(t, err)
	assert.Equal(t, expectedLink, link)
}

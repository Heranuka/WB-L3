package cache

import (
	"context"
	"shortener/internal/config"

	"encoding/json"
	"time"

	"github.com/wb-go/wbf/redis"
)

type CacheService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
}

type RedisService struct {
	client *redis.Client
}

func NewRedisService(cfg *config.Config) *RedisService {
	client := redis.New(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DBRedis)
	return &RedisService{
		client: client,
	}
}

func (s *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {

	return nil
}

func (s *RedisService) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := s.client.Get(ctx, key)
	if err != nil {
		return err
	}
	if strPtr, ok := dest.(*string); ok {
		*strPtr = val
		return nil
	}
	return json.Unmarshal([]byte(val), dest)
}

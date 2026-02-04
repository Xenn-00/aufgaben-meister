package cache

import (
	"context"
	"time"

	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redis *redis.Client) *RedisCache {
	return &RedisCache{client: redis}
}

func (r *RedisCache) Get(ctx context.Context, key string) (*any, *app_errors.AppError) {
	return utils.GetCacheData[any](ctx, r.client, key)
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) *app_errors.AppError {

	return utils.SetCacheData(ctx, r.client, key, &value, ttl)
}

func (r *RedisCache) Del(ctx context.Context, key string) error {
	return utils.DeleteCacheData(ctx, r.client, key)
}

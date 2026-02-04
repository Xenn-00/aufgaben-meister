package cache

import (
	"context"
	"time"

	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
)

type Cache interface {
	Get(ctx context.Context, key string) (*any, *app_errors.AppError)
	Set(ctx context.Context, key string, value any, ttl time.Duration) *app_errors.AppError
	Del(ctx context.Context, key string) error
}

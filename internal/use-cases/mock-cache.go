package use_cases

import (
	"context"
	"time"

	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
)

type MockCache struct {
	GetFn func(ctx context.Context, key string) (*any, *app_errors.AppError)
	SetFn func(ctx context.Context, key string, val *any, ttl time.Duration) *app_errors.AppError
	DelFn func(ctx context.Context, key string) error

	GetCalled int
	SetCalled int
	DelCalled int
}

func (m *MockCache) Get(ctx context.Context, key string) (*any, *app_errors.AppError) {
	m.GetCalled++
	return m.GetFn(ctx, key)
}

func (m *MockCache) Set(ctx context.Context, key string, val any, ttl time.Duration) *app_errors.AppError {
	m.SetCalled++
	return m.SetFn(ctx, key, &val, ttl)
}

func (m *MockCache) Del(ctx context.Context, key string) error {
	m.DelCalled++
	return m.DelFn(ctx, key)
}

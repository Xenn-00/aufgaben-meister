package auth_repo

import (
	"context"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
)

// AuthRepoContract reicht die Methoden für das AuthRepo weiter.
type AuthRepoContract interface {
	CountUsers(ctx context.Context, filter entity.UserCountFilter) (int64, *app_errors.AppError)
	SaveUsers(ctx context.Context, model entity.UserEntity) (string, string, *app_errors.AppError)
	FindByEmail(ctx context.Context, email string) (*entity.UserEntity, *app_errors.AppError)
	FindByUsername(ctx context.Context, username string) (*entity.UserEntity, *app_errors.AppError)
}

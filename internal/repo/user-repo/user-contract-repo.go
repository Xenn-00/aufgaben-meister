package user_repo

import (
	"context"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/jackc/pgx/v5"
)

type UserRepoContract interface {
	FindByUserID(ctx context.Context, userID string) (*entity.UserEntity, *app_errors.AppError)
	IsUnderOneProject(ctx context.Context, reqUserID, userID string) (bool, *app_errors.AppError)
	FindUserWithProjects(ctx context.Context, userID string) (*entity.UserWithProject, *app_errors.AppError)
	UpdateSelfProfileTx(ctx context.Context, tx pgx.Tx, userID string, model entity.UserUpdate) (*entity.UserEntity, *app_errors.AppError)
	DeactivateSelfUser(ctx context.Context, tx pgx.Tx, userID string) (bool, *app_errors.AppError)
}

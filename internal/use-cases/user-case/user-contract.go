package user_case

import (
	"context"

	user_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/user-dto"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
)

type UserServiceContract interface {
	UserSelfProfile(ctx context.Context, userID string) (*user_dto.UserProfileResponse, *app_errors.AppError)
	UserProfileById(ctx context.Context, req user_dto.ParamGetUserByID, viewerUserID string) (*user_dto.UserProfileResponse, *app_errors.AppError)
	UpdateSelfProfile(ctx context.Context, req user_dto.UpdateSelfProfileRequest, userID string) (*user_dto.UserProfileResponse, *app_errors.AppError)
	DeactivateSelfUser(ctx context.Context, req user_dto.DeactivateSelfUserRequest, userID string) *app_errors.AppError
}

package auth_case

import (
	"context"

	auth_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/auth-dto"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
)

// AuthServiceContract reicht die Methoden für den AuthService weiter.
type AuthServiceContract interface {
	RegisterUser(ctx context.Context, req auth_dto.RegisterUserRequest) (*auth_dto.RegisterUserResponse, *app_errors.AppError)
	LoginUser(ctx context.Context, req auth_dto.LoginUserRequest, loginMeta auth_dto.LoginMetadata) (*auth_dto.LoginUserResponse, *app_errors.AppError)
	LogoutUser(ctx context.Context, sessionID string) *app_errors.AppError
	ListAllUserDevices(ctx context.Context, userID string) (*[]auth_dto.ListAllUserDevicesResponse, *app_errors.AppError)
	LogoutAllDevices(ctx context.Context, userID string) *app_errors.AppError
}

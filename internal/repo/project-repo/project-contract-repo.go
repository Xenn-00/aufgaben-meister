package project_repo

import (
	"context"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/jackc/pgx/v5"
)

type ProjectRepoContract interface {
	InsertNewProject(ctx context.Context, tx pgx.Tx, modelProject *entity.ProjectEntity) (*entity.ProjectEntity, *app_errors.AppError)
	InsertNewProjectMember(ctx context.Context, tx pgx.Tx, projectID, userID string, role entity.UserRole) *app_errors.AppError
	GetSelfProject(ctx context.Context, userID string) ([]entity.ProjectSelf, *app_errors.AppError)
	GetUserRoleInProject(ctx context.Context, userID, projectID string) (string, *app_errors.AppError)
	GetProjectByID(ctx context.Context, projectID string) (*entity.ProjectEntity, *app_errors.AppError)
	GetUsersByIds(ctx context.Context, userIDs []string) (map[string]bool, *app_errors.AppError)
	GetProjectMember(ctx context.Context, projectID string) ([]entity.ProjectMember, *app_errors.AppError)
	GetProjectMemberUserIDs(ctx context.Context, projectID string) (map[string]bool, *app_errors.AppError)
	IsProjectExist(ctx context.Context, projectID string) (bool, *app_errors.AppError)
	GetPendingInvitations(ctx context.Context, projectID string) (map[string]bool, *app_errors.AppError)
	GetInvitationInfo(ctx context.Context, invitationID string) (*entity.InvitationInfo, *app_errors.AppError)
	GetInvitationProjectByID(ctx context.Context, invitationID string) (*entity.ProjectInvitationEntity, *app_errors.AppError)
	BatchInsertProjectInvitation(ctx context.Context, tx pgx.Tx, invs []entity.ProjectInvitationEntity) *app_errors.AppError
	UpdateUserInvitationState(ctx context.Context, tx pgx.Tx, invitationID, status string) *app_errors.AppError
}

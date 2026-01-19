package project_repo

import (
	"context"
	"time"

	project_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/project-dto"
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
	GetInvitationProjectByIDWithTx(ctx context.Context, tx pgx.Tx, invitationID string) (*entity.ProjectInvitationEntity, *app_errors.AppError)
	GetInvitationProjectByID(ctx context.Context, invitationID string) (*entity.ProjectInvitationEntity, *app_errors.AppError)
	GetUserPendingInvitations(ctx context.Context, userID string) ([]entity.ProjectInvitationEntity, *app_errors.AppError)
	BatchInsertProjectInvitation(ctx context.Context, tx pgx.Tx, invs []entity.ProjectInvitationEntity) *app_errors.AppError
	AcceptUserInvitationState(ctx context.Context, tx pgx.Tx, invitationID, status string) *app_errors.AppError
	RejectUserInvitationState(ctx context.Context, tx pgx.Tx, invitationID, status string) *app_errors.AppError
	RevokePendingInvitations(ctx context.Context, tx pgx.Tx, projectID string, targetUserIDs []string) ([]string, *app_errors.AppError)
	RevokeAcceptedMembers(ctx context.Context, tx pgx.Tx, projectID string, targetUserIDs []string) ([]string, *app_errors.AppError)
	RotateTokenInvitation(ctx context.Context, tx pgx.Tx, invitationID string, tokenHash string, expiration time.Time) *app_errors.AppError
	ListInvitations(ctx context.Context, projectID string, filters *project_dto.FilterProjectInvitation) ([]entity.ProjectInvitationEntity, *app_errors.AppError)
	ListInvitationsExpire(ctx context.Context, tx pgx.Tx) ([]string, *app_errors.AppError)
	UpdateInvitationsExpire(ctx context.Context, tx pgx.Tx, invitationIDs []string) *app_errors.AppError
}

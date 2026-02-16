package project_case

import (
	"context"

	"github.com/Xenn-00/aufgaben-meister/internal/dtos"
	project_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/project-dto"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
)

type ProjectServiceContract interface {
	CreateNewProject(ctx context.Context, req *project_dto.CreateNewProjectRequest, userID string) (*project_dto.CreateNewProjectResponse, *app_errors.AppError)
	GetSelfProject(ctx context.Context, userID string) ([]*project_dto.SelfProjectResponse, *app_errors.AppError)
	GetProjectDetail(ctx context.Context, projectID, userID string) (*project_dto.GetProjectDetailResponse, *app_errors.AppError)
	InviteProjectMember(ctx context.Context, projectID, userID string, req *project_dto.InviteProjectMemberRequest) (*project_dto.InviteProjectMemberResponse, *app_errors.AppError)
	AcceptInvitationProject(ctx context.Context, req *project_dto.InvitationQueryRequest, userID string) (*project_dto.InvitationMemberAccepted, *app_errors.AppError)
	GetSelfInvitationPending(ctx context.Context, userID string) ([]project_dto.SelfProjectInvitationResponse, *app_errors.AppError)
	RejectProjectInvitation(ctx context.Context, invitationID, userID string) (*project_dto.RejectProjectInvitationResponse, *app_errors.AppError)
	RevokeProjectInvitations(ctx context.Context, projectID, userID string, req *project_dto.RevokeProjectMemberRequest) (*project_dto.RevokeProjectMemberResponse, *app_errors.AppError)
	ResendProjectInvitations(ctx context.Context, invitationID, userID string) *app_errors.AppError
	GetInvitationsInProject(ctx context.Context, projectID, userID string, filters project_dto.FilterProjectInvitation) ([]*project_dto.InvitationsInProjectResponse, *dtos.CursorPaginationMeta, *app_errors.AppError)
}

package project_case

import (
	"context"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/abstraction/tx"
	project_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/project-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/stretchr/testify/mock"
)

type MockProjectRepo struct {
	mock.Mock
}

func (m *MockProjectRepo) InsertNewProject(ctx context.Context, t tx.Tx, modelProject *entity.ProjectEntity) (*entity.ProjectEntity, *app_errors.AppError) {
	args := m.Called(ctx, t, modelProject)
	return args.Get(0).(*entity.ProjectEntity), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) InsertNewProjectMember(ctx context.Context, t tx.Tx, projectID, userID string, role entity.UserRole) *app_errors.AppError {
	args := m.Called(ctx, t, projectID, userID, role)
	return args.Get(0).(*app_errors.AppError)
}

func (m *MockProjectRepo) GetSelfProject(ctx context.Context, userID string) ([]entity.ProjectSelf, *app_errors.AppError) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).([]entity.ProjectSelf), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) GetUserRoleInProject(ctx context.Context, userID, projectID string) (string, *app_errors.AppError) {
	args := m.Called(ctx, userID, projectID)
	return args.String(0), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) GetProjectByID(ctx context.Context, projectID string) (*entity.ProjectEntity, *app_errors.AppError) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).(*entity.ProjectEntity), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) GetUsersByIds(ctx context.Context, userIDs []string) (map[string]bool, *app_errors.AppError) {
	args := m.Called(ctx, userIDs)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).(map[string]bool), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) GetProjectMember(ctx context.Context, projectID string) ([]entity.ProjectMember, *app_errors.AppError) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).([]entity.ProjectMember), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) GetProjectMemberUserIDs(ctx context.Context, projectID string) (map[string]bool, *app_errors.AppError) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).(map[string]bool), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) IsProjectExist(ctx context.Context, projectID string) (bool, *app_errors.AppError) {
	args := m.Called(ctx, projectID)
	return args.Bool(0), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) GetPendingInvitations(ctx context.Context, projectID string) (map[string]bool, *app_errors.AppError) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).(map[string]bool), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) GetInvitationInfo(ctx context.Context, invitationID string) (*entity.InvitationInfo, *app_errors.AppError) {
	args := m.Called(ctx, invitationID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).(*entity.InvitationInfo), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) GetInvitationProjectByIDWithTx(ctx context.Context, t tx.Tx, invitationID string) (*entity.ProjectInvitationEntity, *app_errors.AppError) {
	args := m.Called(ctx, t, invitationID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).(*entity.ProjectInvitationEntity), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) GetInvitationProjectByID(ctx context.Context, invitationID string) (*entity.ProjectInvitationEntity, *app_errors.AppError) {
	args := m.Called(ctx, invitationID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).(*entity.ProjectInvitationEntity), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) GetUserPendingInvitations(ctx context.Context, userID string) ([]entity.ProjectInvitationEntity, *app_errors.AppError) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).([]entity.ProjectInvitationEntity), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) BatchInsertProjectInvitation(ctx context.Context, t tx.Tx, invs []entity.ProjectInvitationEntity) *app_errors.AppError {
	args := m.Called(ctx, t, invs)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*app_errors.AppError)
}

func (m *MockProjectRepo) AcceptUserInvitationState(ctx context.Context, t tx.Tx, invitationID, status string) *app_errors.AppError {
	args := m.Called(ctx, t, invitationID, status)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*app_errors.AppError)
}

func (m *MockProjectRepo) RejectUserInvitationState(ctx context.Context, t tx.Tx, invitationID, status string) *app_errors.AppError {
	args := m.Called(ctx, t, invitationID, status)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*app_errors.AppError)
}

func (m *MockProjectRepo) RevokePendingInvitations(ctx context.Context, t tx.Tx, projectID string, targetUserIDs []string) ([]string, *app_errors.AppError) {
	args := m.Called(ctx, t, projectID, targetUserIDs)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).([]string), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) RevokeAcceptedMembers(ctx context.Context, t tx.Tx, projectID string, targetUserIDs []string) ([]string, *app_errors.AppError) {
	args := m.Called(ctx, t, projectID, targetUserIDs)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).([]string), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) RotateTokenInvitation(ctx context.Context, t tx.Tx, invitationID string, tokenHash string, expiration time.Time) *app_errors.AppError {
	args := m.Called(ctx, t, invitationID, tokenHash, expiration)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*app_errors.AppError)
}

func (m *MockProjectRepo) ListInvitations(ctx context.Context, projectID string, filters *project_dto.FilterProjectInvitation) ([]entity.ProjectInvitationEntity, *app_errors.AppError) {
	args := m.Called(ctx, projectID, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).([]entity.ProjectInvitationEntity), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) ListInvitationsExpire(ctx context.Context, t tx.Tx) ([]string, *app_errors.AppError) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*app_errors.AppError)
	}
	return args.Get(0).([]string), args.Get(1).(*app_errors.AppError)
}

func (m *MockProjectRepo) UpdateInvitationsExpire(ctx context.Context, t tx.Tx, invitationIDs []string) *app_errors.AppError {
	args := m.Called(ctx, t, invitationIDs)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*app_errors.AppError)
}

package project_case

import (
	"context"
	"testing"

	project_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/project-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	use_cases "github.com/Xenn-00/aufgaben-meister/internal/use-cases"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test happy path - Meister successfully invites new users
func TestInviteProjectMember_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
	taskQueue := new(use_cases.MockTaskQueue)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
		taskQueue: taskQueue,
	}

	projectID := "project-1"
	meisterID := "meister-1"
	userIDs := []string{"user-1", "user-2"}

	req := &project_dto.InviteProjectMemberRequest{
		UserIDs: userIDs,
	}

	// Setup expectations
	repo.On("IsProjectExist", ctx, projectID).Return(true, (*app_errors.AppError)(nil))

	// GetUserRoleInProject - must be MEISTER
	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	// GetProjectMemberUserIDs - empty (no existing members)
	memberSet := map[string]bool{}
	repo.On("GetProjectMemberUserIDs", ctx, projectID).Return(memberSet, (*app_errors.AppError)(nil))

	// GetPendingInvitations - empty (no pending invites)
	inviteSet := map[string]bool{}
	repo.On("GetPendingInvitations", ctx, projectID).Return(inviteSet, (*app_errors.AppError)(nil))

	// GetUsersByIds - both users exist
	userExistSet := map[string]bool{
		"user-1": true,
		"user-2": true,
	}
	repo.On("GetUsersByIds", ctx, userIDs).Return(userExistSet, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// BatchInsertProjectInvitation - check that 2 invitations are created
	repo.On("BatchInsertProjectInvitation", ctx, tx, mock.MatchedBy(func(invs []entity.ProjectInvitationEntity) bool {
		return len(invs) == 2 &&
			invs[0].ProjectID == projectID &&
			invs[1].ProjectID == projectID &&
			invs[0].InvitedBy == meisterID &&
			invs[1].InvitedBy == meisterID &&
			invs[0].Role == entity.MITARBEITER &&
			invs[1].Role == entity.MITARBEITER
	})).Return((*app_errors.AppError)(nil))

	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	taskQueue.On("EnqueueSendInvitationEmail", mock.Anything).Return(nil).Times(2)

	resp, err := service.InviteProjectMember(ctx, projectID, meisterID, req)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Invited, 2)
	assert.Contains(t, resp.Invited, "user-1")
	assert.Contains(t, resp.Invited, "user-2")
	assert.Len(t, resp.SkippedUsers, 0)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
	taskQueue.AssertExpectations(t)
}

// Test 2: User is not MEISTER (forbidden)
func TestInviteProjectMember_UserNotMeister(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	projectID := "project-1"
	mitarbeiterID := "mitarbeiter-1"
	userIDs := []string{"user-1"}

	req := &project_dto.InviteProjectMemberRequest{
		UserIDs: userIDs,
	}

	// Setup expectations
	repo.On("IsProjectExist", ctx, projectID).Return(true, (*app_errors.AppError)(nil))

	// GetUserRoleInProject - user is MITARBEITER
	mitarbeiterRole := string(entity.MITARBEITER)
	repo.On("GetUserRoleInProject", ctx, mitarbeiterID, projectID).Return(mitarbeiterRole, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.InviteProjectMember(ctx, projectID, mitarbeiterID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 3: Project does not exist
func TestInviteProjectMember_ProjectNotExist(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	projectID := "project-999"
	meisterID := "meister-1"
	userIDs := []string{"user-1"}

	req := &project_dto.InviteProjectMemberRequest{
		UserIDs: userIDs,
	}

	// Setup expectations
	// IsProjectExist returns error
	notExistError := app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
	repo.On("IsProjectExist", ctx, projectID).Return(false, notExistError)

	// Execute
	resp, err := service.InviteProjectMember(ctx, projectID, meisterID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, notExistError, err)

	repo.AssertExpectations(t)
}

// Test 4: All users are skipped (various skip reasons)
func TestInviteProjectMember_AllUsersSkipped(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	projectID := "project-1"
	meisterID := "meister-1"
	userIDs := []string{"user-1", "user-2", "user-3", "user-4"}

	req := &project_dto.InviteProjectMemberRequest{
		UserIDs: userIDs,
	}

	// Setup expectations
	repo.On("IsProjectExist", ctx, projectID).Return(true, (*app_errors.AppError)(nil))

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	// user-2 is already a member
	memberSet := map[string]bool{
		"user-2": true,
	}
	repo.On("GetProjectMemberUserIDs", ctx, projectID).Return(memberSet, (*app_errors.AppError)(nil))

	// user-3 already has pending invitation
	inviteSet := map[string]bool{
		"user-3": true,
	}
	repo.On("GetPendingInvitations", ctx, projectID).Return(inviteSet, (*app_errors.AppError)(nil))

	// user-1 doesn't exist, user-4 doesn't exist
	userExistSet := map[string]bool{
		"user-1": false, // doesn't exist
		"user-2": true,
		"user-3": true,
		"user-4": false, // doesn't exist
	}
	repo.On("GetUsersByIds", ctx, userIDs).Return(userExistSet, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.InviteProjectMember(ctx, projectID, meisterID, req)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Invited, 0) // No one invited
	assert.Len(t, resp.SkippedUsers, 4)

	// Verify skip reasons
	skipReasons := make(map[string]string)
	for _, skipped := range resp.SkippedUsers {
		skipReasons[skipped.UserID] = skipped.Reason
	}

	assert.Equal(t, "user_not_found", skipReasons["user-1"])
	assert.Equal(t, "invitation.already_member", skipReasons["user-2"])
	assert.Equal(t, "invitation.already_invited", skipReasons["user-3"])
	assert.Equal(t, "user_not_found", skipReasons["user-4"])

	repo.AssertExpectations(t)
}

// Test 5: Mixed scenario - some invited, some skipped
func TestInviteProjectMember_MixedInviteAndSkip(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
	taskQueue := new(use_cases.MockTaskQueue)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
		taskQueue: taskQueue,
	}

	projectID := "project-1"
	meisterID := "meister-1"
	userIDs := []string{"user-1", "user-2", "user-3"}

	req := &project_dto.InviteProjectMemberRequest{
		UserIDs: userIDs,
	}

	// Setup expectations
	repo.On("IsProjectExist", ctx, projectID).Return(true, (*app_errors.AppError)(nil))

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	// user-2 is already a member
	memberSet := map[string]bool{
		"user-2": true,
	}
	repo.On("GetProjectMemberUserIDs", ctx, projectID).Return(memberSet, (*app_errors.AppError)(nil))

	// No pending invitations
	inviteSet := map[string]bool{}
	repo.On("GetPendingInvitations", ctx, projectID).Return(inviteSet, (*app_errors.AppError)(nil))

	// user-1 and user-3 exist, user-2 exists
	userExistSet := map[string]bool{
		"user-1": true,
		"user-2": true,
		"user-3": true,
	}
	repo.On("GetUsersByIds", ctx, userIDs).Return(userExistSet, (*app_errors.AppError)(nil))

	// Transaction - only user-1 and user-3 should be invited
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	repo.On("BatchInsertProjectInvitation", ctx, tx, mock.MatchedBy(func(invs []entity.ProjectInvitationEntity) bool {
		return len(invs) == 2 // Only 2 users invited
	})).Return((*app_errors.AppError)(nil))

	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// EnqueueSendInvitationEmail - called twice (for user-1 and user-3)
	taskQueue.On("EnqueueSendInvitationEmail", mock.Anything).Return(nil).Times(2)

	// Execute
	resp, err := service.InviteProjectMember(ctx, projectID, meisterID, req)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Invited, 2)
	assert.Contains(t, resp.Invited, "user-1")
	assert.Contains(t, resp.Invited, "user-3")
	assert.NotContains(t, resp.Invited, "user-2")

	assert.Len(t, resp.SkippedUsers, 1)
	assert.Equal(t, "user-2", resp.SkippedUsers[0].UserID)
	assert.Equal(t, "invitation.already_member", resp.SkippedUsers[0].Reason)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
	taskQueue.AssertExpectations(t)
}

// Test 6: Transaction commit fails
func TestInviteProjectMember_TransactionCommitFails(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	projectID := "project-1"
	meisterID := "meister-1"
	userIDs := []string{"user-1"}

	req := &project_dto.InviteProjectMemberRequest{
		UserIDs: userIDs,
	}

	// Setup expectations
	repo.On("IsProjectExist", ctx, projectID).Return(true, (*app_errors.AppError)(nil))

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	memberSet := map[string]bool{}
	repo.On("GetProjectMemberUserIDs", ctx, projectID).Return(memberSet, (*app_errors.AppError)(nil))

	inviteSet := map[string]bool{}
	repo.On("GetPendingInvitations", ctx, projectID).Return(inviteSet, (*app_errors.AppError)(nil))

	userExistSet := map[string]bool{
		"user-1": true,
	}
	repo.On("GetUsersByIds", ctx, userIDs).Return(userExistSet, (*app_errors.AppError)(nil))

	// Transaction
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	repo.On("BatchInsertProjectInvitation", ctx, tx, mock.Anything).Return((*app_errors.AppError)(nil))

	// Commit fails
	commitError := app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "commit_failed", nil)
	tx.On("Commit", ctx).Return(commitError)
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.InviteProjectMember(ctx, projectID, meisterID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, err.Code)
	assert.Equal(t, app_errors.ErrInternal, err.Type)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

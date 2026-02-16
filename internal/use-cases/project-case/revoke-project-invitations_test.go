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
)

// Test 1: Happy path - Revoke both pending and accepted members
func TestRevokeProjectInvitations_Success_MixedRevoke(t *testing.T) {
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
	userIDs := []string{"user-1", "user-2", "user-3", "user-4"}

	req := &project_dto.RevokeProjectMemberRequest{
		UserIDs: userIDs,
	}

	// Setup expectations
	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// Strategy A: user-1 and user-2 had pending invitations
	pendingRevoked := []string{"user-1", "user-2"}
	repo.On("RevokePendingInvitations", ctx, tx, projectID, userIDs).Return(pendingRevoked, (*app_errors.AppError)(nil))

	// Strategy B: user-3 and user-4 were accepted members
	acceptedRevoked := []string{"user-3", "user-4"}
	repo.On("RevokeAcceptedMembers", ctx, tx, projectID, userIDs).Return(acceptedRevoked, (*app_errors.AppError)(nil))

	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.RevokeProjectInvitations(ctx, projectID, meisterID, req)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Revoked, 4)
	assert.Contains(t, resp.Revoked, "user-1")
	assert.Contains(t, resp.Revoked, "user-2")
	assert.Contains(t, resp.Revoked, "user-3")
	assert.Contains(t, resp.Revoked, "user-4")

	// Verify revoke reasons
	revokeReasons := make(map[string]string)
	for _, u := range resp.RevokedUsers {
		revokeReasons[u.UserID] = u.Reason
	}

	assert.Equal(t, "invitation_revoked_before_acceptance", revokeReasons["user-1"])
	assert.Equal(t, "invitation_revoked_before_acceptance", revokeReasons["user-2"])
	assert.Equal(t, "membership_revoked_after_acceptance", revokeReasons["user-3"])
	assert.Equal(t, "membership_revoked_after_acceptance", revokeReasons["user-4"])

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 2: Happy path - Revoke only pending invitations (Strategy A only)
func TestRevokeProjectInvitations_Success_PendingOnly(t *testing.T) {
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
	userIDs := []string{"user-1", "user-2"}

	req := &project_dto.RevokeProjectMemberRequest{
		UserIDs: userIDs,
	}

	// Setup expectations
	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// Strategy A: both users had pending invitations
	pendingRevoked := []string{"user-1", "user-2"}
	repo.On("RevokePendingInvitations", ctx, tx, projectID, userIDs).Return(pendingRevoked, (*app_errors.AppError)(nil))

	// Strategy B: no accepted members to revoke
	repo.On("RevokeAcceptedMembers", ctx, tx, projectID, userIDs).Return([]string{}, (*app_errors.AppError)(nil))

	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.RevokeProjectInvitations(ctx, projectID, meisterID, req)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Revoked, 2)

	for _, u := range resp.RevokedUsers {
		assert.Equal(t, "invitation_revoked_before_acceptance", u.Reason)
	}

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 3: Happy path - Revoke only accepted members (Strategy B only)
func TestRevokeProjectInvitations_Success_AcceptedOnly(t *testing.T) {
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
	userIDs := []string{"user-1", "user-2"}

	req := &project_dto.RevokeProjectMemberRequest{
		UserIDs: userIDs,
	}

	// Setup expectations
	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// Strategy A: no pending invitations
	repo.On("RevokePendingInvitations", ctx, tx, projectID, userIDs).Return([]string{}, (*app_errors.AppError)(nil))

	// Strategy B: both users were accepted members
	acceptedRevoked := []string{"user-1", "user-2"}
	repo.On("RevokeAcceptedMembers", ctx, tx, projectID, userIDs).Return(acceptedRevoked, (*app_errors.AppError)(nil))

	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.RevokeProjectInvitations(ctx, projectID, meisterID, req)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Revoked, 2)

	for _, u := range resp.RevokedUsers {
		assert.Equal(t, "membership_revoked_after_acceptance", u.Reason)
	}

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 4: User is not MEISTER
func TestRevokeProjectInvitations_UserNotMeister(t *testing.T) {
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

	req := &project_dto.RevokeProjectMemberRequest{
		UserIDs: userIDs,
	}

	// User is MITARBEITER
	mitarbeiterRole := string(entity.MITARBEITER)
	repo.On("GetUserRoleInProject", ctx, mitarbeiterID, projectID).Return(mitarbeiterRole, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.RevokeProjectInvitations(ctx, projectID, mitarbeiterID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 5: RevokePendingInvitations repo call fails
func TestRevokeProjectInvitations_RevokePendingFails(t *testing.T) {
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
	userIDs := []string{"user-1", "user-2"}

	req := &project_dto.RevokeProjectMemberRequest{
		UserIDs: userIDs,
	}

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// RevokePendingInvitations fails
	dbError := app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "database_error", nil)
	repo.On("RevokePendingInvitations", ctx, tx, projectID, userIDs).Return(([]string)(nil), dbError)

	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.RevokeProjectInvitations(ctx, projectID, meisterID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, dbError, err)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 6: RevokeAcceptedMembers repo call fails
func TestRevokeProjectInvitations_RevokeAcceptedFails(t *testing.T) {
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
	userIDs := []string{"user-1", "user-2"}

	req := &project_dto.RevokeProjectMemberRequest{
		UserIDs: userIDs,
	}

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// Strategy A: success
	pendingRevoked := []string{"user-1"}
	repo.On("RevokePendingInvitations", ctx, tx, projectID, userIDs).Return(pendingRevoked, (*app_errors.AppError)(nil))

	// Strategy B: fails
	dbError := app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "database_error", nil)
	repo.On("RevokeAcceptedMembers", ctx, tx, projectID, userIDs).Return(([]string)(nil), dbError)

	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.RevokeProjectInvitations(ctx, projectID, meisterID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, dbError, err)

	// Rollback is called (via defer) so Strategy A changes are also rolled back
	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

package project_case

import (
	"context"
	"testing"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	use_cases "github.com/Xenn-00/aufgaben-meister/internal/use-cases"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Test 1: Happy path - Successfully reject invitation
func TestRejectProjectInvitation_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	invitationID := "invitation-1"
	userID := "user-1"
	projectID := "project-1"

	// Setup expectations
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// GetInvitationProjectByIDWithTx
	invitation := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: userID,
		InvitedBy:     "meister-1",
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING, // Still pending
		TokenHash:     "some-hash",
		ExpiresAt:     time.Now().Add(24 * time.Hour), // Not expired
	}

	repo.On("GetInvitationProjectByIDWithTx", ctx, tx, invitationID).Return(invitation, (*app_errors.AppError)(nil))

	// RejectUserInvitationState
	repo.On("RejectUserInvitationState", ctx, tx, invitationID, string(entity.REJECTED)).Return((*app_errors.AppError)(nil))

	// Commit
	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// GetInvitationProjectByID (after commit)
	rejectedInvitation := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: userID,
		Role:          entity.MITARBEITER,
		Status:        entity.REJECTED, // Status updated
		ExpiresAt:     invitation.ExpiresAt,
	}

	repo.On("GetInvitationProjectByID", ctx, invitationID).Return(rejectedInvitation, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.RejectProjectInvitation(ctx, invitationID, userID)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, invitationID, resp.ID)
	assert.Equal(t, projectID, resp.ProjectID)
	assert.Equal(t, string(entity.REJECTED), resp.Status)
	assert.Equal(t, invitation.ExpiresAt, resp.ExpiresAt)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 2: UserID mismatch (trying to reject someone else's invitation)
func TestRejectProjectInvitation_UserIDMismatch(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	invitationID := "invitation-1"
	invitedUserID := "user-1"
	differentUserID := "user-2" // Different user trying to reject
	projectID := "project-1"

	// Setup expectations
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	invitation := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: invitedUserID, // Invited user-1
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	repo.On("GetInvitationProjectByIDWithTx", ctx, tx, invitationID).Return(invitation, (*app_errors.AppError)(nil))

	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute with different user
	resp, err := service.RejectProjectInvitation(ctx, invitationID, differentUserID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 3: Invitation expired
func TestRejectProjectInvitation_InvitationExpired(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	invitationID := "invitation-1"
	userID := "user-1"
	projectID := "project-1"

	// Setup expectations
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	invitation := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: userID,
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING,
		ExpiresAt:     time.Now().Add(-24 * time.Hour), // EXPIRED (24 hours ago)
	}

	repo.On("GetInvitationProjectByIDWithTx", ctx, tx, invitationID).Return(invitation, (*app_errors.AppError)(nil))

	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.RejectProjectInvitation(ctx, invitationID, userID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict", err.MessageKey)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 4: Invitation already accepted (not pending)
func TestRejectProjectInvitation_AlreadyAccepted(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	invitationID := "invitation-1"
	userID := "user-1"
	projectID := "project-1"

	// Setup expectations
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	invitation := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: userID,
		Role:          entity.MITARBEITER,
		Status:        entity.ACCEPTED, // Already accepted!
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	repo.On("GetInvitationProjectByIDWithTx", ctx, tx, invitationID).Return(invitation, (*app_errors.AppError)(nil))

	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.RejectProjectInvitation(ctx, invitationID, userID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict", err.MessageKey)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 5: Invitation not found
func TestRejectProjectInvitation_InvitationNotFound(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	invitationID := "invitation-999"
	userID := "user-1"

	// Setup expectations
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// Invitation not found
	notFoundError := app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "invitation_not_found", nil)
	repo.On("GetInvitationProjectByIDWithTx", ctx, tx, invitationID).Return((*entity.ProjectInvitationEntity)(nil), notFoundError)

	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.RejectProjectInvitation(ctx, invitationID, userID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, notFoundError, err)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 6: Transaction commit fails
func TestRejectProjectInvitation_CommitFails(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	invitationID := "invitation-1"
	userID := "user-1"
	projectID := "project-1"

	// Setup expectations
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	invitation := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: userID,
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	repo.On("GetInvitationProjectByIDWithTx", ctx, tx, invitationID).Return(invitation, (*app_errors.AppError)(nil))

	repo.On("RejectUserInvitationState", ctx, tx, invitationID, string(entity.REJECTED)).Return((*app_errors.AppError)(nil))

	// Commit fails
	commitError := app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "commit_failed", nil)
	tx.On("Commit", ctx).Return(commitError)
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.RejectProjectInvitation(ctx, invitationID, userID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, err.Code)
	assert.Equal(t, app_errors.ErrInternal, err.Type)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

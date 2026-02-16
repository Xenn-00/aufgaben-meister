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
	"github.com/stretchr/testify/mock"
)

// Test 1: Happy path - MEISTER successfully resends invitation
func TestResendProjectInvitations_Success(t *testing.T) {
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

	invitationID := "invitation-1"
	meisterID := "meister-1"
	projectID := "project-1"

	// Setup expectations
	inv := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: "user-1",
		InvitedBy:     meisterID,
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING,
		TokenHash:     "old-token-hash",
		ExpiresAt:     time.Now().Add(24 * time.Hour), // Not expired
	}

	repo.On("GetInvitationProjectByID", ctx, invitationID).Return(inv, (*app_errors.AppError)(nil))

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	// Transaction
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// RotateTokenInvitation - new token hash and extended expiry
	repo.On("RotateTokenInvitation", ctx, tx, invitationID, mock.AnythingOfType("string"), mock.MatchedBy(func(t time.Time) bool {
		return t.After(time.Now()) // New expiry must be in the future
	})).Return((*app_errors.AppError)(nil))

	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// EnqueueSendInvitationEmail
	taskQueue.On("EnqueueSendInvitationEmail", mock.Anything).Return(nil)

	// Execute
	err := service.ResendProjectInvitations(ctx, invitationID, meisterID)

	// Assert
	assert.Nil(t, err)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
	taskQueue.AssertExpectations(t)
}

// Test 2: Invitation not found
func TestResendProjectInvitations_InvitationNotFound(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	invitationID := "invitation-999"
	meisterID := "meister-1"

	// Invitation not found
	notFoundError := app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "invitation_not_found", nil)
	repo.On("GetInvitationProjectByID", ctx, invitationID).Return((*entity.ProjectInvitationEntity)(nil), notFoundError)

	// Execute
	err := service.ResendProjectInvitations(ctx, invitationID, meisterID)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, notFoundError, err)

	repo.AssertExpectations(t)
}

// Test 3: User is not MEISTER
func TestResendProjectInvitations_UserNotMeister(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	invitationID := "invitation-1"
	mitarbeiterID := "mitarbeiter-1"
	projectID := "project-1"

	inv := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: "user-1",
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	repo.On("GetInvitationProjectByID", ctx, invitationID).Return(inv, (*app_errors.AppError)(nil))

	// User is MITARBEITER, not MEISTER
	mitarbeiterRole := string(entity.MITARBEITER)
	repo.On("GetUserRoleInProject", ctx, mitarbeiterID, projectID).Return(mitarbeiterRole, (*app_errors.AppError)(nil))

	// Execute
	err := service.ResendProjectInvitations(ctx, invitationID, mitarbeiterID)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 4: Invitation already expired
func TestResendProjectInvitations_InvitationExpired(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	invitationID := "invitation-1"
	meisterID := "meister-1"
	projectID := "project-1"

	inv := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: "user-1",
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING,
		ExpiresAt:     time.Now().Add(-24 * time.Hour), // EXPIRED
	}

	repo.On("GetInvitationProjectByID", ctx, invitationID).Return(inv, (*app_errors.AppError)(nil))

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	// Execute
	err := service.ResendProjectInvitations(ctx, invitationID, meisterID)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 5: Invitation status is not Pending (already accepted/rejected)
func TestResendProjectInvitations_InvitationNotPending(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	invitationID := "invitation-1"
	meisterID := "meister-1"
	projectID := "project-1"

	inv := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: "user-1",
		Role:          entity.MITARBEITER,
		Status:        entity.ACCEPTED, // Already accepted!
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	repo.On("GetInvitationProjectByID", ctx, invitationID).Return(inv, (*app_errors.AppError)(nil))

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	// Execute
	err := service.ResendProjectInvitations(ctx, invitationID, meisterID)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 6: Transaction commit fails
func TestResendProjectInvitations_CommitFails(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	invitationID := "invitation-1"
	meisterID := "meister-1"
	projectID := "project-1"

	inv := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: "user-1",
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	repo.On("GetInvitationProjectByID", ctx, invitationID).Return(inv, (*app_errors.AppError)(nil))

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	repo.On("RotateTokenInvitation", ctx, tx, invitationID, mock.AnythingOfType("string"), mock.MatchedBy(func(t time.Time) bool {
		return t.After(time.Now())
	})).Return((*app_errors.AppError)(nil))

	// Commit fails
	commitError := app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "commit_failed", nil)
	tx.On("Commit", ctx).Return(commitError)
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	err := service.ResendProjectInvitations(ctx, invitationID, meisterID)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, err.Code)
	assert.Equal(t, app_errors.ErrInternal, err.Type)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

package project_case

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	project_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/project-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	use_cases "github.com/Xenn-00/aufgaben-meister/internal/use-cases"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Test 1: Happy path - Successfully accept invitation
func TestAcceptInvitationProject_Success(t *testing.T) {
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
	rawToken := "valid-token-12345678"

	// Calculate token hash (same as service does)
	tokenHash := sha256.Sum256([]byte(rawToken))
	expectedHash := hex.EncodeToString(tokenHash[:])

	req := &project_dto.InvitationQueryRequest{
		InvitationID: invitationID,
		Token:        rawToken,
	}

	// Setup expectations
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// GetInvitationProjectByIDWithTx
	invitation := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: userID,
		InvitedBy:     "meister-1",
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING,
		TokenHash:     expectedHash,                   // Correct token hash
		ExpiresAt:     time.Now().Add(24 * time.Hour), // Not expired
	}

	repo.On("GetInvitationProjectByIDWithTx", ctx, tx, invitationID).Return(invitation, (*app_errors.AppError)(nil))

	// AcceptUserInvitationState
	repo.On("AcceptUserInvitationState", ctx, tx, invitationID, string(entity.ACCEPTED)).Return((*app_errors.AppError)(nil))

	// InsertNewProjectMember
	repo.On("InsertNewProjectMember", ctx, tx, projectID, userID, entity.MITARBEITER).Return((*app_errors.AppError)(nil))

	// Commit
	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// GetProjectByID (after commit)
	project := &entity.ProjectEntity{
		ID:   projectID,
		Name: "Test Project",
		Type: entity.CORPORATE,
	}

	repo.On("GetProjectByID", ctx, projectID).Return(project, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.AcceptInvitationProject(ctx, req, userID)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, projectID, resp.ID)
	assert.Equal(t, "Test Project", resp.Name)
	assert.Equal(t, string(entity.MITARBEITER), resp.Role)
	assert.NotZero(t, resp.AcceptedAt)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 2: Invalid token (token mismatch)
func TestAcceptInvitationProject_InvalidToken(t *testing.T) {
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
	wrongToken := "wrong-token-12345678"
	correctToken := "correct-token-123456"

	// Calculate hash for the CORRECT token (stored in DB)
	correctTokenHash := sha256.Sum256([]byte(correctToken))
	storedHash := hex.EncodeToString(correctTokenHash[:])

	req := &project_dto.InvitationQueryRequest{
		InvitationID: invitationID,
		Token:        wrongToken, // Wrong token!
	}

	// Setup expectations
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	invitation := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: userID,
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING,
		TokenHash:     storedHash, // Hash of correct token
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	repo.On("GetInvitationProjectByIDWithTx", ctx, tx, invitationID).Return(invitation, (*app_errors.AppError)(nil))

	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.AcceptInvitationProject(ctx, req, userID)

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

// Test 3: UserID mismatch (trying to accept someone else's invitation)
func TestAcceptInvitationProject_UserIDMismatch(t *testing.T) {
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
	differentUserID := "user-2" // Different user trying to accept
	projectID := "project-1"
	rawToken := "valid-token-12345678"

	tokenHash := sha256.Sum256([]byte(rawToken))
	expectedHash := hex.EncodeToString(tokenHash[:])

	req := &project_dto.InvitationQueryRequest{
		InvitationID: invitationID,
		Token:        rawToken,
	}

	// Setup expectations
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	invitation := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: invitedUserID, // Invited user-1
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING,
		TokenHash:     expectedHash,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	repo.On("GetInvitationProjectByIDWithTx", ctx, tx, invitationID).Return(invitation, (*app_errors.AppError)(nil))

	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute with different user
	resp, err := service.AcceptInvitationProject(ctx, req, differentUserID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 4: Invitation expired
func TestAcceptInvitationProject_InvitationExpired(t *testing.T) {
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
	rawToken := "valid-token-12345678"

	tokenHash := sha256.Sum256([]byte(rawToken))
	expectedHash := hex.EncodeToString(tokenHash[:])

	req := &project_dto.InvitationQueryRequest{
		InvitationID: invitationID,
		Token:        rawToken,
	}

	// Setup expectations
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	invitation := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: userID,
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING,
		TokenHash:     expectedHash,
		ExpiresAt:     time.Now().Add(-24 * time.Hour), // EXPIRED (24 hours ago)
	}

	repo.On("GetInvitationProjectByIDWithTx", ctx, tx, invitationID).Return(invitation, (*app_errors.AppError)(nil))

	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.AcceptInvitationProject(ctx, req, userID)

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
func TestAcceptInvitationProject_InvitationNotFound(t *testing.T) {
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
	rawToken := "valid-token-12345678"

	req := &project_dto.InvitationQueryRequest{
		InvitationID: invitationID,
		Token:        rawToken,
	}

	// Setup expectations
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// Invitation not found
	notFoundError := app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "invitation_not_found", nil)
	repo.On("GetInvitationProjectByIDWithTx", ctx, tx, invitationID).Return((*entity.ProjectInvitationEntity)(nil), notFoundError)

	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.AcceptInvitationProject(ctx, req, userID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, notFoundError, err)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 6: Transaction commit fails
func TestAcceptInvitationProject_CommitFails(t *testing.T) {
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
	rawToken := "valid-token-12345678"

	tokenHash := sha256.Sum256([]byte(rawToken))
	expectedHash := hex.EncodeToString(tokenHash[:])

	req := &project_dto.InvitationQueryRequest{
		InvitationID: invitationID,
		Token:        rawToken,
	}

	// Setup expectations
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	invitation := &entity.ProjectInvitationEntity{
		ID:            invitationID,
		ProjectID:     projectID,
		InvitedUserID: userID,
		Role:          entity.MITARBEITER,
		Status:        entity.PENDING,
		TokenHash:     expectedHash,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	repo.On("GetInvitationProjectByIDWithTx", ctx, tx, invitationID).Return(invitation, (*app_errors.AppError)(nil))

	repo.On("AcceptUserInvitationState", ctx, tx, invitationID, string(entity.ACCEPTED)).Return((*app_errors.AppError)(nil))

	repo.On("InsertNewProjectMember", ctx, tx, projectID, userID, entity.MITARBEITER).Return((*app_errors.AppError)(nil))

	// Commit fails
	commitError := app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "commit_failed", nil)
	tx.On("Commit", ctx).Return(commitError)
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.AcceptInvitationProject(ctx, req, userID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, err.Code)
	assert.Equal(t, app_errors.ErrInternal, err.Type)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

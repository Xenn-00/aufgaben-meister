package project_case

import (
	"context"
	"testing"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Test 1: Happy path - User has multiple pending invitations
func TestGetSelfInvitationPending_Success_MultipleInvitations(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{
		repo: repo,
	}

	userID := "user-1"

	// Setup expectations
	projectName1 := "Project Alpha"
	projectName2 := "Project Beta"

	invitations := []entity.ProjectInvitationEntity{
		{
			ID:          "inv-1",
			ProjectID:   "project-1",
			ProjectName: &projectName1,
			Role:        entity.MITARBEITER,
			Status:      entity.PENDING,
			ExpiresAt:   time.Now().Add(5 * 24 * time.Hour),
			InvitedBy:   "meister-1",
		},
		{
			ID:          "inv-2",
			ProjectID:   "project-2",
			ProjectName: &projectName2,
			Role:        entity.MITARBEITER,
			Status:      entity.PENDING,
			ExpiresAt:   time.Now().Add(3 * 24 * time.Hour),
			InvitedBy:   "meister-2",
		},
	}

	repo.On("GetUserPendingInvitations", ctx, userID).Return(invitations, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.GetSelfInvitationPending(ctx, userID)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 2)

	// Verify first invitation
	assert.Equal(t, "inv-1", resp[0].ID)
	assert.Equal(t, "project-1", resp[0].ProjectID)
	assert.Equal(t, "Project Alpha", resp[0].ProjectName)
	assert.Equal(t, string(entity.MITARBEITER), resp[0].Role)
	assert.Equal(t, string(entity.PENDING), resp[0].Status)
	assert.Equal(t, "meister-1", resp[0].InvitedBy)

	// Verify second invitation
	assert.Equal(t, "inv-2", resp[1].ID)
	assert.Equal(t, "project-2", resp[1].ProjectID)
	assert.Equal(t, "Project Beta", resp[1].ProjectName)

	repo.AssertExpectations(t)
}

// Test 2: User has no pending invitations (empty list)
func TestGetSelfInvitationPending_Success_NoInvitations(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{
		repo: repo,
	}

	userID := "user-1"

	// Setup expectations - empty list
	emptyInvitations := []entity.ProjectInvitationEntity{}

	repo.On("GetUserPendingInvitations", ctx, userID).Return(emptyInvitations, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.GetSelfInvitationPending(ctx, userID)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 0) // Empty but not nil

	repo.AssertExpectations(t)
}

// Test 3: User has single pending invitation
func TestGetSelfInvitationPending_Success_SingleInvitation(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{
		repo: repo,
	}

	userID := "user-1"

	// Setup expectations
	projectName := "Solo Project"
	invitations := []entity.ProjectInvitationEntity{
		{
			ID:          "inv-1",
			ProjectID:   "project-1",
			ProjectName: &projectName,
			Role:        entity.MITARBEITER,
			Status:      entity.PENDING,
			ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
			InvitedBy:   "meister-1",
		},
	}

	repo.On("GetUserPendingInvitations", ctx, userID).Return(invitations, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.GetSelfInvitationPending(ctx, userID)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, "inv-1", resp[0].ID)
	assert.Equal(t, "Solo Project", resp[0].ProjectName)

	repo.AssertExpectations(t)
}

// Test 4: Repository error
func TestGetSelfInvitationPending_RepositoryError(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{
		repo: repo,
	}

	userID := "user-1"

	// Setup expectations - repository returns error
	dbError := app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "database_error", nil)
	repo.On("GetUserPendingInvitations", ctx, userID).Return(([]entity.ProjectInvitationEntity)(nil), dbError)

	// Execute
	resp, err := service.GetSelfInvitationPending(ctx, userID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, dbError, err)

	repo.AssertExpectations(t)
}

// Test 5: User has invitations with different roles
func TestGetSelfInvitationPending_Success_DifferentRoles(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{
		repo: repo,
	}

	userID := "user-1"

	// Setup expectations - invitations with different roles
	projectName1 := "Project One"
	projectName2 := "Project Two"

	invitations := []entity.ProjectInvitationEntity{
		{
			ID:          "inv-1",
			ProjectID:   "project-1",
			ProjectName: &projectName1,
			Role:        entity.MITARBEITER, // Regular member role
			Status:      entity.PENDING,
			ExpiresAt:   time.Now().Add(5 * 24 * time.Hour),
			InvitedBy:   "meister-1",
		},
		{
			ID:          "inv-2",
			ProjectID:   "project-2",
			ProjectName: &projectName2,
			Role:        entity.MEISTER, // Master role (edge case, but possible)
			Status:      entity.PENDING,
			ExpiresAt:   time.Now().Add(3 * 24 * time.Hour),
			InvitedBy:   "admin-1",
		},
	}

	repo.On("GetUserPendingInvitations", ctx, userID).Return(invitations, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.GetSelfInvitationPending(ctx, userID)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 2)
	assert.Equal(t, string(entity.MITARBEITER), resp[0].Role)
	assert.Equal(t, string(entity.MEISTER), resp[1].Role)

	repo.AssertExpectations(t)
}

// Test 6: User has invitations with different expiry times (some about to expire)
func TestGetSelfInvitationPending_Success_DifferentExpiryTimes(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{
		repo: repo,
	}

	userID := "user-1"

	// Setup expectations - invitations with different expiry times
	projectName1 := "Project Urgent"
	projectName2 := "Project Normal"

	soonExpiry := time.Now().Add(2 * time.Hour)        // Expires soon
	normalExpiry := time.Now().Add(5 * 24 * time.Hour) // Normal expiry

	invitations := []entity.ProjectInvitationEntity{
		{
			ID:          "inv-1",
			ProjectID:   "project-1",
			ProjectName: &projectName1,
			Role:        entity.MITARBEITER,
			Status:      entity.PENDING,
			ExpiresAt:   soonExpiry, // About to expire
			InvitedBy:   "meister-1",
		},
		{
			ID:          "inv-2",
			ProjectID:   "project-2",
			ProjectName: &projectName2,
			Role:        entity.MITARBEITER,
			Status:      entity.PENDING,
			ExpiresAt:   normalExpiry, // Normal expiry
			InvitedBy:   "meister-2",
		},
	}

	repo.On("GetUserPendingInvitations", ctx, userID).Return(invitations, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.GetSelfInvitationPending(ctx, userID)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 2)

	// Verify expiry times are preserved
	assert.Equal(t, soonExpiry, resp[0].ExpiresAt)
	assert.Equal(t, normalExpiry, resp[1].ExpiresAt)

	// Client can use this to show urgent invitations
	assert.True(t, resp[0].ExpiresAt.Before(resp[1].ExpiresAt))

	repo.AssertExpectations(t)
}

package project_case

import (
	"context"
	"testing"
	"time"

	project_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/project-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Test 1: Happy path - MEISTER gets invitations with cursor pagination (has more)
func TestGetInvitationsInProject_Success_HasMore(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{repo: repo}

	projectID := "project-1"
	meisterID := "meister-1"
	limit := 2

	filters := project_dto.FilterProjectInvitation{
		Limit: limit,
	}

	// Setup expectations
	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	project := &entity.ProjectEntity{
		ID:   projectID,
		Name: "Test Project",
	}
	repo.On("GetProjectByID", ctx, projectID).Return(project, (*app_errors.AppError)(nil))

	// Return limit+1 rows to signal hasMore
	now := time.Now()
	rows := []entity.ProjectInvitationEntity{
		{ID: "inv-1", InvitedUserID: "user-1", Status: entity.PENDING, ExpiresAt: now.Add(5 * 24 * time.Hour), CreatedAt: now.Add(-3 * time.Hour)},
		{ID: "inv-2", InvitedUserID: "user-2", Status: entity.PENDING, ExpiresAt: now.Add(6 * 24 * time.Hour), CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "inv-3", InvitedUserID: "user-3", Status: entity.PENDING, ExpiresAt: now.Add(7 * 24 * time.Hour), CreatedAt: now.Add(-1 * time.Hour)}, // extra row
	}

	repo.On("ListInvitations", ctx, projectID, &filters).Return(rows, (*app_errors.AppError)(nil))

	// Execute
	data, cursor, err := service.GetInvitationsInProject(ctx, projectID, meisterID, filters)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.NotNil(t, cursor)
	assert.Len(t, data, 2) // Trimmed to limit
	assert.True(t, cursor.HasMore)
	assert.NotNil(t, cursor.NextCursor) // Has next cursor
	assert.Equal(t, limit, cursor.Limit)
	assert.Equal(t, "inv-1", data[0].InvitationID)
	assert.Equal(t, "inv-2", data[1].InvitationID)
	assert.Equal(t, "Test Project", data[0].ProjectName)

	repo.AssertExpectations(t)
}

// Test 2: Happy path - Last page (no more data)
func TestGetInvitationsInProject_Success_NoMore(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{repo: repo}

	projectID := "project-1"
	meisterID := "meister-1"
	limit := 10

	filters := project_dto.FilterProjectInvitation{
		Limit: limit,
	}

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	project := &entity.ProjectEntity{ID: projectID, Name: "Test Project"}
	repo.On("GetProjectByID", ctx, projectID).Return(project, (*app_errors.AppError)(nil))

	// Return less than limit rows (last page)
	now := time.Now()
	rows := []entity.ProjectInvitationEntity{
		{ID: "inv-1", InvitedUserID: "user-1", Status: entity.PENDING, ExpiresAt: now.Add(5 * 24 * time.Hour), CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "inv-2", InvitedUserID: "user-2", Status: entity.ACCEPTED, ExpiresAt: now.Add(6 * 24 * time.Hour), CreatedAt: now.Add(-1 * time.Hour)},
	}

	repo.On("ListInvitations", ctx, projectID, &filters).Return(rows, (*app_errors.AppError)(nil))

	// Execute
	data, cursor, err := service.GetInvitationsInProject(ctx, projectID, meisterID, filters)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Len(t, data, 2)
	assert.False(t, cursor.HasMore)
	assert.Nil(t, cursor.NextCursor) // No next cursor
	assert.Equal(t, limit, cursor.Limit)

	repo.AssertExpectations(t)
}

// Test 3: User is not MEISTER
func TestGetInvitationsInProject_UserNotMeister(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{repo: repo}

	projectID := "project-1"
	mitarbeiterID := "mitarbeiter-1"

	filters := project_dto.FilterProjectInvitation{Limit: 10}

	mitarbeiterRole := string(entity.MITARBEITER)
	repo.On("GetUserRoleInProject", ctx, mitarbeiterID, projectID).Return(mitarbeiterRole, (*app_errors.AppError)(nil))

	// Execute
	data, cursor, err := service.GetInvitationsInProject(ctx, projectID, mitarbeiterID, filters)

	// Assert
	assert.Nil(t, data)
	assert.Nil(t, cursor)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 4: Expired filter used without Pending status (invalid filter combo)
func TestGetInvitationsInProject_ExpiredFilterWithoutPendingStatus(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{repo: repo}

	projectID := "project-1"
	meisterID := "meister-1"

	// filters.Expired is set but Status is not "Pending"
	expiredFilter := true
	acceptedStatus := "Accepted"
	filters := project_dto.FilterProjectInvitation{
		Limit:   10,
		Expired: &expiredFilter,
		Status:  &acceptedStatus, // Wrong! Should be "Pending"
	}

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	project := &entity.ProjectEntity{ID: projectID, Name: "Test Project"}
	repo.On("GetProjectByID", ctx, projectID).Return(project, (*app_errors.AppError)(nil))

	// Execute
	data, cursor, err := service.GetInvitationsInProject(ctx, projectID, meisterID, filters)

	// Assert
	assert.Nil(t, data)
	assert.Nil(t, cursor)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusBadRequest, err.Code)
	assert.Equal(t, app_errors.ErrInvalidQuery, err.Type)
	assert.Equal(t, "invitation.expired_only_valid_for_pending", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 5: Default limit applied when limit is 0
func TestGetInvitationsInProject_DefaultLimit(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{repo: repo}

	projectID := "project-1"
	meisterID := "meister-1"

	// Limit is 0, should default to 20
	filters := project_dto.FilterProjectInvitation{
		Limit: 0,
	}

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	project := &entity.ProjectEntity{ID: projectID, Name: "Test Project"}
	repo.On("GetProjectByID", ctx, projectID).Return(project, (*app_errors.AppError)(nil))

	now := time.Now()
	rows := []entity.ProjectInvitationEntity{
		{ID: "inv-1", InvitedUserID: "user-1", Status: entity.PENDING, ExpiresAt: now.Add(5 * 24 * time.Hour), CreatedAt: now},
	}

	// Expect filters with limit=20 (default applied before repo call)
	expectedFilters := project_dto.FilterProjectInvitation{Limit: 20}
	repo.On("ListInvitations", ctx, projectID, &expectedFilters).Return(rows, (*app_errors.AppError)(nil))

	// Execute
	data, cursor, err := service.GetInvitationsInProject(ctx, projectID, meisterID, filters)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 20, cursor.Limit) // Default limit applied

	repo.AssertExpectations(t)
}

// Test 6: Limit exceeds max (100) - capped to 100
func TestGetInvitationsInProject_LimitCappedAt100(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{repo: repo}

	projectID := "project-1"
	meisterID := "meister-1"

	// Limit is 999, should be capped to 100
	filters := project_dto.FilterProjectInvitation{
		Limit: 999,
	}

	meisterRole := string(entity.MEISTER)
	repo.On("GetUserRoleInProject", ctx, meisterID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	project := &entity.ProjectEntity{ID: projectID, Name: "Test Project"}
	repo.On("GetProjectByID", ctx, projectID).Return(project, (*app_errors.AppError)(nil))

	now := time.Now()
	rows := []entity.ProjectInvitationEntity{
		{ID: "inv-1", InvitedUserID: "user-1", Status: entity.PENDING, ExpiresAt: now.Add(5 * 24 * time.Hour), CreatedAt: now},
	}

	// Expect filters with limit=100 (capped before repo call)
	expectedFilters := project_dto.FilterProjectInvitation{Limit: 100}
	repo.On("ListInvitations", ctx, projectID, &expectedFilters).Return(rows, (*app_errors.AppError)(nil))

	// Execute
	data, cursor, err := service.GetInvitationsInProject(ctx, projectID, meisterID, filters)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 100, cursor.Limit) // Capped to 100

	repo.AssertExpectations(t)
}

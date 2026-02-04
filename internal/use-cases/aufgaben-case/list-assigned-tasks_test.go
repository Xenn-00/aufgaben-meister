package aufgaben_case

import (
	"context"
	"testing"
	"time"

	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/stretchr/testify/assert"
)

// Test 1: Happy path
func TestListAssignedTasks_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
	}

	userID := "user-1"

	status := "In_Progress"
	projectID := "project-1"
	priority := "Medium"

	filter := &aufgaben_dto.AssignedAufgabenFilter{
		Status:    &status,
		Limit:     2,
		Priority:  &priority,
		ProjectID: &projectID,
	}

	// since projectID was define, than we need to mock project member

	repo.On("CheckProjectMember", ctx, *filter.ProjectID, userID).Return(true, (*app_errors.AppError)(nil))

	description := "Important"
	dueDate := time.Now().Add(24 * time.Hour)
	assignedTasks := []entity.AssignedAufgaben{
		{
			ID:          "task-1",
			ProjectName: "Test project",
			Title:       "title task",
			Description: &description,
			Status:      entity.AufgabenInProgress,
			Priority:    entity.PriorityMedium,
			DueDate:     dueDate,
		},
		{
			ID:          "task-2",
			ProjectName: "Test project 2",
			Title:       "title task 2",
			Description: &description,
			Status:      entity.AufgabenInProgress,
			Priority:    entity.PriorityMedium,
			DueDate:     dueDate,
		},
		{
			ID:          "task-3",
			ProjectName: "Test project 3",
			Title:       "title task 3",
			Description: &description,
			Status:      entity.AufgabenInProgress,
			Priority:    entity.PriorityMedium,
			DueDate:     dueDate,
		},
	}

	repo.On("ListAssignedTasks", ctx, userID, filter).Return([]entity.AssignedAufgaben(assignedTasks), (*app_errors.AppError)(nil))

	resp, c, err := service.ListAssignedTasks(ctx, userID, filter)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, c)

	assert.Equal(t, "Test project", resp[0].ProjectName)
	assert.Equal(t, "task-1", resp[0].AufgabenID)
	assert.Equal(t, true, c.HasMore)
	assert.Equal(t, 2, c.Limit)
	assert.Equal(t, "task-2", c.NextCursor)

	repo.AssertExpectations(t)
}

// Test 2: not a member of project
func TestListAssignedTasks_NotProjectMember(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
	}

	userID := "user-1"

	status := "In_Progress"
	projectID := "project-1"
	priority := "Medium"

	filter := &aufgaben_dto.AssignedAufgabenFilter{
		Status:    &status,
		Limit:     2,
		Priority:  &priority,
		ProjectID: &projectID,
	}

	// since projectID was define, than we need to mock project member
	repo.On("CheckProjectMember", ctx, *filter.ProjectID, userID).Return(false, (*app_errors.AppError)(nil))

	resp, c, err := service.ListAssignedTasks(ctx, userID, filter)

	assert.NotNil(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, c)

	assert.Equal(t, app_errors.ErrForbidden, err.Type)

	repo.AssertExpectations(t)
}

// Test 3: invalid cursor
func TestListAssignedTasks_InvalidCursor(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
	}

	userID := "user-1"

	status := "In_Progress"
	projectID := "project-1"
	priority := "Medium"
	cursor := "not-uuid"

	filter := &aufgaben_dto.AssignedAufgabenFilter{
		Status:    &status,
		Limit:     2,
		Priority:  &priority,
		ProjectID: &projectID,
		Cursor:    &cursor,
	}

	// since projectID was define, than we need to mock project member
	repo.On("CheckProjectMember", ctx, *filter.ProjectID, userID).Return(true, (*app_errors.AppError)(nil))

	resp, c, err := service.ListAssignedTasks(ctx, userID, filter)

	assert.NotNil(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, c)

	assert.Equal(t, app_errors.ErrInvalidQuery, err.Type)

	repo.AssertExpectations(t)
}

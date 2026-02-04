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

// Test Happy path
func TestListTasksProject_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
	}

	userID := "user-1"
	projectID := "project-1"
	assigneeID := "user-2"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	status := "In_Progress"
	filters := &aufgaben_dto.AufgabenListFilter{
		Status:     &status,
		Limit:      10,
		Page:       1,
		AssigneeID: &assigneeID,
	}

	repo.On("CheckProjectMember", ctx, projectID, assigneeID).Return(true, (*app_errors.AppError)(nil)) // <- for assignee

	taskDescription := "Important"
	dueDate := time.Now()

	r := []entity.AufgabenEntity{
		{
			ID:          "task-1",
			Title:       "Test Task",
			Description: &taskDescription,
			Status:      entity.AufgabenInProgress,
			Priority:    entity.PriorityMedium,
			AssigneeID:  &assigneeID,
			DueDate:     &dueDate,
		},
	}
	repo.On("ListTasks", ctx, projectID, filters).Return(([]entity.AufgabenEntity)(r), (*app_errors.AppError)(nil))
	repo.On("CountTasks", ctx, projectID).Return(1, (*app_errors.AppError)(nil))

	resp, paging, err := service.ListTasksProject(ctx, userID, projectID, *filters)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, paging)

	assert.Equal(t, "Test Task", resp[0].Title)
	assert.Equal(t, "user-2", *resp[0].AssigneeID)
	assert.Equal(t, 1, paging.Page)
	assert.Equal(t, 10, paging.Limit)
	assert.Equal(t, 1, paging.Total)
	assert.Equal(t, 1, paging.TotalPages)

	repo.AssertExpectations(t)
}

// Test performer not project member
func TestListTasksProject_PerformerNotProjectMember(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
	}

	userID := "user-1"
	projectID := "project-1"
	assigneeID := "user-2"

	status := "In_Progress"
	filters := &aufgaben_dto.AufgabenListFilter{
		Status:     &status,
		Limit:      10,
		Page:       1,
		AssigneeID: &assigneeID,
	}

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(false, (*app_errors.AppError)(nil))

	resp, paging, err := service.ListTasksProject(ctx, userID, projectID, *filters)

	assert.NotNil(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, paging)

	assert.Equal(t, app_errors.ErrForbidden, err.Type)

	repo.AssertExpectations(t)
}

// Test assigneeID given in filter is not a project member but performer is project member
func TestListtasksProject_AssigneeNotProjectMemberButPerformerYes(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
	}

	userID := "user-1"
	projectID := "project-1"
	assigneeID := "user-from-another-project"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil)) // <- for performer

	status := "In_Progress"
	filters := &aufgaben_dto.AufgabenListFilter{
		Status:     &status,
		Limit:      10,
		Page:       1,
		AssigneeID: &assigneeID,
	}

	repo.On("CheckProjectMember", ctx, projectID, assigneeID).Return(false, (*app_errors.AppError)(nil)) // <- for assignee

	resp, paging, err := service.ListTasksProject(ctx, userID, projectID, *filters)

	assert.NotNil(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, paging)

	assert.Equal(t, app_errors.ErrForbidden, err.Type)

	repo.AssertExpectations(t)
}

// Test when tasks return an empty list
func TestListTasksProject_ReturnEmptyList(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
	}

	userID := "user-1"
	projectID := "project-1"
	assigneeID := "user-2"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	status := "In_Progress"
	filters := &aufgaben_dto.AufgabenListFilter{
		Status:     &status,
		Limit:      10,
		Page:       1,
		AssigneeID: &assigneeID,
	}

	repo.On("CheckProjectMember", ctx, projectID, assigneeID).Return(true, (*app_errors.AppError)(nil)) // <- for assignee

	r := []entity.AufgabenEntity{}
	repo.On("ListTasks", ctx, projectID, filters).Return(([]entity.AufgabenEntity)(r), (*app_errors.AppError)(nil))
	repo.On("CountTasks", ctx, projectID).Return(1, (*app_errors.AppError)(nil))

	resp, paging, err := service.ListTasksProject(ctx, userID, projectID, *filters)

	assert.Nil(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, paging)

	assert.Equal(t, 0, len(resp))
	assert.Equal(t, 1, paging.Page)
	assert.Equal(t, 10, paging.Limit)
	assert.Equal(t, 1, paging.Total)
	assert.Equal(t, 1, paging.TotalPages)

	repo.AssertExpectations(t)
}

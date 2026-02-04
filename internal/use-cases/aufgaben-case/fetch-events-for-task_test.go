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

func TestFetchEventsForTask_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	filter := &aufgaben_dto.AufgabenEventFilter{
		Limit: 2,
	}

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	dueDate := time.Now().Add(48 * time.Hour)
	projectName := "Test Project"
	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		ProjectID:   projectID,
		ProjectName: &projectName,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  func() *string { s := "old-user"; return &s }(),
		DueDate:     &dueDate,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	assignmenEvents := []entity.AssignmentEventEntity{
		{
			ID:               "event-1",
			AufgabenID:       taskID,
			ActorID:          userID,
			Action:           entity.ActionAssign,
			Note:             func() *string { s := "assign-event"; return &s }(),
			TargetAssigneeID: nil,
			ReasonCode:       nil,
			ReasonText:       nil,
			CreatedAt:        time.Now().Add(1 * time.Hour),
		},
		{
			ID:               "event-2",
			AufgabenID:       taskID,
			ActorID:          userID,
			Action:           entity.ActionDueDateUpdate,
			Note:             func() *string { s := "update-duedate-event"; return &s }(),
			TargetAssigneeID: nil,
			ReasonCode:       nil,
			ReasonText:       nil,
			CreatedAt:        time.Now().Add(45 * time.Minute),
		},
		{
			ID:               "event-3",
			AufgabenID:       taskID,
			ActorID:          userID,
			Action:           entity.ActionComplete,
			Note:             func() *string { s := "complete"; return &s }(),
			TargetAssigneeID: nil,
			ReasonCode:       nil,
			ReasonText:       nil,
			CreatedAt:        time.Now().Add(10 * time.Minute),
		},
	}

	repo.On("ListEventsForTask", ctx, taskID, filter).Return([]entity.AssignmentEventEntity(assignmenEvents), (*app_errors.AppError)(nil))

	resp, c, err := service.FetchEventsForTask(ctx, userID, projectID, taskID, filter)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, c)

	assert.Equal(t, true, c.HasMore)
	assert.Equal(t, 2, c.Limit)
	assert.Equal(t, "event-2", c.NextCursor)

	repo.AssertExpectations(t)
}

func TestFetchEventsForTask_NotProjectMember(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	filter := &aufgaben_dto.AufgabenEventFilter{
		Limit: 1,
	}

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(false, (*app_errors.AppError)(nil))

	resp, c, err := service.FetchEventsForTask(ctx, userID, projectID, taskID, filter)

	assert.NotNil(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, c)

	assert.Equal(t, app_errors.ErrForbidden, err.Type)

	repo.AssertExpectations(t)
}

func TestFetchEventsForTask_InvalidCursor(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"
	cursor := "not-uuid"

	filter := &aufgaben_dto.AufgabenEventFilter{
		Limit:  1,
		Cursor: &cursor,
	}

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	dueDate := time.Now().Add(48 * time.Hour)
	projectName := "Test Project"
	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		ProjectID:   projectID,
		ProjectName: &projectName,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  func() *string { s := "old-user"; return &s }(),
		DueDate:     &dueDate,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	resp, c, err := service.FetchEventsForTask(ctx, userID, projectID, taskID, filter)

	assert.NotNil(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, c)

	assert.Equal(t, app_errors.ErrInvalidQuery, err.Type)

	repo.AssertExpectations(t)
}

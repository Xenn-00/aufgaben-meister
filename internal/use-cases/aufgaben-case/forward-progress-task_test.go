package aufgaben_case

import (
	"context"
	"testing"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test Happy path
func TestForwardProgressTask_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	tx := new(MockTx)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	// Setup expectations
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &userID, // Task is assigned to userID
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	completedAt := time.Now()
	completedTask := &entity.CompleteTaskEntity{
		ID:          taskID,
		Status:      entity.AufgabenDone,
		Priority:    entity.PriorityHigh,
		AssigneeID:  userID,
		CreatedBy:   "creator-1",
		CompletedAt: completedAt,
	}

	repo.On("ForwardProgress", ctx, tx, taskID).Return(completedTask, (*app_errors.AppError)(nil))

	repo.On("InsertAssignmentEvent", ctx, tx, mock.Anything).Return((*app_errors.AppError)(nil))

	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForwardProgressTask(ctx, userID, projectID, taskID)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, taskID, resp.AufgabenID)
	assert.Equal(t, string(entity.AufgabenDone), resp.Status)
	assert.Equal(t, completedAt, resp.CompletedAt)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test User is not a project member
func TestForwardProgressTask_UserNotProjectMember(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	// User is not a project member
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(false, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForwardProgressTask(ctx, userID, projectID, taskID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, 403, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// Task is archived
func TestForwardProgressTask_TaskArchived(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	archivedTime := time.Now().Add(-48 * time.Hour)
	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &userID,
		ArchivedAt:  &archivedTime, // Task is archived
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForwardProgressTask(ctx, userID, projectID, taskID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, 409, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_unavailable", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 4: Task status is already Done
func TestForwardProgressTask_TaskAlreadyDone(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenDone, // Already done
		Priority:    entity.PriorityHigh,
		AssigneeID:  &userID,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForwardProgressTask(ctx, userID, projectID, taskID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, 409, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_is_archive_or_done", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 5: User is not the task assignee
func TestForwardProgressTask_NotTaskAssignee(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "user-1"
	differentUserID := "user-2" // Different user assigned to task
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &differentUserID, // Assigned to different user
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForwardProgressTask(ctx, userID, projectID, taskID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, 403, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden.not_task_assignee", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 6: Task has no assignee (unassigned task)
func TestForwardProgressTask_TaskNotAssigned(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenTodo,
		Priority:    entity.PriorityHigh,
		AssigneeID:  nil, // No assignee
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForwardProgressTask(ctx, userID, projectID, taskID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, 403, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden.not_task_assignee", err.MessageKey)

	repo.AssertExpectations(t)
}

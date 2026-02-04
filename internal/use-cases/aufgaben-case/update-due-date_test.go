package aufgaben_case

import (
	"context"
	"testing"
	"time"

	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test 1: Happy path - Successfully update due date
func TestUpdateDueDate_Success(t *testing.T) {
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

	newDueDate := time.Now().Add(72 * time.Hour)
	req := &aufgaben_dto.UpdateDueDateRequest{
		DueDate: newDueDate,
	}

	// Setup expectations
	// getTaskAndVerifyMember calls
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	oldDueDate := time.Now().Add(48 * time.Hour)
	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		DueDate:     &oldDueDate,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// Transaction
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// UpdateDueDate
	repo.On("UpdateDueDate", ctx, tx, taskID, newDueDate).Return(&newDueDate, (*app_errors.AppError)(nil))

	// InsertAssignmentEvent
	repo.On("InsertAssignmentEvent", ctx, tx, mock.Anything).Return((*app_errors.AppError)(nil))

	// Commit
	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.UpdateDueDate(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, taskID, resp.AufgabenID)
	assert.Equal(t, newDueDate, resp.DueDate)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 2: User is not a project member
func TestUpdateDueDate_UserNotProjectMember(t *testing.T) {
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

	newDueDate := time.Now().Add(72 * time.Hour)
	req := &aufgaben_dto.UpdateDueDateRequest{
		DueDate: newDueDate,
	}

	// User is not a project member
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(false, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.UpdateDueDate(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 3: Task not found
func TestUpdateDueDate_TaskNotFound(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-999"

	newDueDate := time.Now().Add(72 * time.Hour)
	req := &aufgaben_dto.UpdateDueDateRequest{
		DueDate: newDueDate,
	}

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	notFoundError := app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "task_not_found", nil)
	repo.On("GetTaskByID", ctx, taskID).Return((*entity.AufgabenEntity)(nil), notFoundError)

	// Execute
	resp, err := service.UpdateDueDate(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, notFoundError, err)

	repo.AssertExpectations(t)
}

// Test 4: Task is archived
func TestUpdateDueDate_TaskArchived(t *testing.T) {
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

	newDueDate := time.Now().Add(72 * time.Hour)
	req := &aufgaben_dto.UpdateDueDateRequest{
		DueDate: newDueDate,
	}

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	archivedTime := time.Now().Add(-48 * time.Hour)
	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		ArchivedAt:  &archivedTime, // Task is archived
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.UpdateDueDate(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_unavailable", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 5: Task status is Done
func TestUpdateDueDate_TaskStatusDone(t *testing.T) {
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

	newDueDate := time.Now().Add(72 * time.Hour)
	req := &aufgaben_dto.UpdateDueDateRequest{
		DueDate: newDueDate,
	}

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenDone, // Status is Done
		Priority:    entity.PriorityHigh,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.UpdateDueDate(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_is_archive_or_done", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 6: Repository UpdateDueDate fails
func TestUpdateDueDate_RepoUpdateFails(t *testing.T) {
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

	newDueDate := time.Now().Add(72 * time.Hour)
	req := &aufgaben_dto.UpdateDueDateRequest{
		DueDate: newDueDate,
	}

	// Setup expectations
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	oldDueDate := time.Now().Add(48 * time.Hour)
	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		DueDate:     &oldDueDate,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// UpdateDueDate fails
	updateError := app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "update_failed", nil)
	repo.On("UpdateDueDate", ctx, tx, taskID, newDueDate).Return((*time.Time)(nil), updateError)

	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.UpdateDueDate(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, updateError, err)

	repo.AssertExpectations(t)
	txManager.AssertExpectations(t)
	tx.AssertExpectations(t)
}

package aufgaben_case

import (
	"context"
	"testing"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	use_cases "github.com/Xenn-00/aufgaben-meister/internal/use-cases"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test 1: Happy path
func TestArchiveTask_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
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
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &userID, // Task is assigned to userID
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// verifyUserRole call
	meisterRole := entity.MEISTER
	repo.On("GetUserRole", ctx, projectID, userID).Return(&meisterRole, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	repo.On("ArchiveTask", ctx, tx, taskID).Return((*app_errors.AppError)(nil))

	repo.On("InsertAssignmentEvent", ctx, tx, mock.Anything).Return((*app_errors.AppError)(nil))

	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	err := service.ArchiveTask(ctx, userID, projectID, taskID)

	assert.Nil(t, err)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 2: User is not a project member
func TestArchiveTask_UserNotProjectMember(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(false, (*app_errors.AppError)(nil))

	// Assert
	err := service.ArchiveTask(ctx, userID, projectID, taskID)

	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// // Test 3: User is not meister
func TestArchiveTask_UserNotMeister(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
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
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &userID, // Task is assigned to userID
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// verifyUserRole call
	meisterRole := entity.MITARBEITER
	repo.On("GetUserRole", ctx, projectID, userID).Return(&meisterRole, (*app_errors.AppError)(nil))

	err := service.ArchiveTask(ctx, userID, projectID, taskID)

	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 4: Task is already archived
func TestArchiveTask_TaskArchived(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
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
		Status:      entity.AufgabenArchived,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &userID, // Task is assigned to userID
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	err := service.ArchiveTask(ctx, userID, projectID, taskID)

	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_is_archive_or_done", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 5: Task is already done
func TestArchiveTask_TaskStatusDone(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
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
		Status:      entity.AufgabenDone,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &userID, // Task is assigned to userID
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	err := service.ArchiveTask(ctx, userID, projectID, taskID)

	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_is_archive_or_done", err.MessageKey)

	repo.AssertExpectations(t)
}

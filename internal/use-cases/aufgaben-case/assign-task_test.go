package aufgaben_case

import (
	"context"
	"testing"
	"time"

	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test Happy path
func TestAssignTask_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	tx := new(MockTx)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	dueDate := time.Now().Add(24 * time.Hour)
	req := &aufgaben_dto.AufgabenAssignRequest{
		DueDate: dueDate,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	// expectation
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important"
	task := &entity.AufgabenEntity{
		ID:          "task-1",
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenTodo,
		Priority:    entity.PriorityMedium,
		AssigneeID:  nil,
		DueDate:     nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return((*entity.AufgabenEntity)(task), (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	assigned := &entity.AssignTaskEntity{
		ID:         taskID,
		Status:     entity.AufgabenInProgress,
		Priority:   entity.PriorityMedium,
		AssigneeID: userID,
		CreatedBy:  userID,
		DueDate:    dueDate,
	}

	repo.On("AssignTask", ctx, tx, projectID, taskID, userID, &dueDate).Return(assigned, (*app_errors.AppError)(nil))

	repo.On("InsertAssignmentEvent", ctx, tx, mock.Anything).Return((*app_errors.AppError)(nil))

	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))

	resp, err := service.AssignTask(ctx, userID, projectID, taskID, req)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, taskID, resp.AufgabenID)
	assert.Equal(t, userID, resp.AssigneeID)
	assert.Equal(t, string(entity.AufgabenInProgress), resp.Status)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test not a project member
func TestAssignTask_UserNotProjectMember(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	dueDate := time.Now().Add(24 * time.Hour)
	req := &aufgaben_dto.AufgabenAssignRequest{
		DueDate: dueDate,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	// User is not a project member
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(false, (*app_errors.AppError)(nil))

	resp, err := service.AssignTask(ctx, userID, projectID, taskID, req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test when task is not found
func TestAssignTask_TaskNotFound(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	dueDate := time.Now().Add(24 * time.Hour)
	req := &aufgaben_dto.AufgabenAssignRequest{
		DueDate: dueDate,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-999"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	notFoundError := app_errors.NewAppError(404, app_errors.ErrNotFound, "task_not_found", nil)
	repo.On("GetTaskByID", ctx, taskID).Return((*entity.AufgabenEntity)(nil), notFoundError)

	resp, err := service.AssignTask(ctx, userID, projectID, taskID, req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, notFoundError, err)

	repo.AssertExpectations(t)
}

// Test when task is already archived (ArchivedAt is not nil)
func TestAssignTask_TaskArchived(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	dueDate := time.Now().Add(24 * time.Hour)
	req := &aufgaben_dto.AufgabenAssignRequest{
		DueDate: dueDate,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	archivedTime := time.Now().Add(-48 * time.Hour)
	taskDescription := "Important"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenTodo,
		Priority:    entity.PriorityMedium,
		AssigneeID:  nil,
		ArchivedAt:  &archivedTime, // Task is archived
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	resp, err := service.AssignTask(ctx, userID, projectID, taskID, req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, 409, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_unavailable", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test when task status is Archived
func TestAssignTask_TaskStatusArchived(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	dueDate := time.Now().Add(24 * time.Hour)
	req := &aufgaben_dto.AufgabenAssignRequest{
		DueDate: dueDate,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenArchived, // Status is Archived
		Priority:    entity.PriorityMedium,
		AssigneeID:  nil,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	resp, err := service.AssignTask(ctx, userID, projectID, taskID, req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, 409, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_is_archive_or_done", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test when task status is Done
func TestAssignTask_TaskStatusDone(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	dueDate := time.Now().Add(24 * time.Hour)
	req := &aufgaben_dto.AufgabenAssignRequest{
		DueDate: dueDate,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenDone, // Status is Done
		Priority:    entity.PriorityMedium,
		AssigneeID:  nil,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	resp, err := service.AssignTask(ctx, userID, projectID, taskID, req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, 409, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_is_archive_or_done", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test when task is already assigned to someone
func TestAssignTask_TaskAlreadyAssigned(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	dueDate := time.Now().Add(24 * time.Hour)
	req := &aufgaben_dto.AufgabenAssignRequest{
		DueDate: dueDate,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	existingAssignee := "user-2"
	taskDescription := "Important"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityMedium,
		AssigneeID:  &existingAssignee, // Already assigned to user-2
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	resp, err := service.AssignTask(ctx, userID, projectID, taskID, req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, 409, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_already_assigned_to_someone", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test when due date is in the past
func TestAssignTask_DueDateInPast(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	pastDate := time.Now().Add(-24 * time.Hour) // Past date
	req := &aufgaben_dto.AufgabenAssignRequest{
		DueDate: pastDate,
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenTodo,
		Priority:    entity.PriorityMedium,
		AssigneeID:  nil,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	resp, err := service.AssignTask(ctx, userID, projectID, taskID, req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, 400, err.Code)
	assert.Equal(t, app_errors.ErrInvalidBody, err.Type)
	assert.Equal(t, "request.invalid_body", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test when due date is exactly now (should fail as it's not "after" now)
func TestAssignTask_DueDateIsNow(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	now := time.Now()
	req := &aufgaben_dto.AufgabenAssignRequest{
		DueDate: now, // Exactly now
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenTodo,
		Priority:    entity.PriorityMedium,
		AssigneeID:  nil,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	resp, err := service.AssignTask(ctx, userID, projectID, taskID, req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, 400, err.Code)
	assert.Equal(t, app_errors.ErrInvalidBody, err.Type)

	repo.AssertExpectations(t)
}

// Test when due date is now + one hour (should fail as it's not "after" now + one hour)
func TestAssignTask_DueDateIsOneHour(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	now := time.Now()
	req := &aufgaben_dto.AufgabenAssignRequest{
		DueDate: now.Add(1 * time.Hour), // now + 1 hour
	}

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenTodo,
		Priority:    entity.PriorityMedium,
		AssigneeID:  nil,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	resp, err := service.AssignTask(ctx, userID, projectID, taskID, req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, 400, err.Code)
	assert.Equal(t, app_errors.ErrInvalidBody, err.Type)

	repo.AssertExpectations(t)
}

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

// Test 1: Happy path - MEISTER successfully force handover task
func TestForceAufgabeHandover_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	tx := new(MockTx)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	meisterID := "meister-1"
	projectID := "project-1"
	taskID := "task-1"
	currentAssignee := "user-1"
	targetID := "user-2"

	req := &aufgaben_dto.ForceAufgabeHandoverRequest{
		TargetID:   targetID,
		Reason:     "Better suited for this task",
		ReasonCode: "Other",
		Note:       "Please take this over",
	}

	// Setup expectations
	// getTaskAndVerifyMember calls
	repo.On("CheckProjectMember", ctx, projectID, meisterID).Return(true, (*app_errors.AppError)(nil))

	dueDate := time.Now().Add(48 * time.Hour)
	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &currentAssignee, // Currently assigned to user-1
		DueDate:     &dueDate,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// verifyUserRole - must be MEISTER
	meisterRole := entity.MEISTER
	repo.On("GetUserRole", ctx, projectID, meisterID).Return(&meisterRole, (*app_errors.AppError)(nil))

	// Transaction
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// AssignTask to new target
	newTask := &entity.AssignTaskEntity{
		ID:         taskID,
		Status:     entity.AufgabenInProgress,
		Priority:   entity.PriorityHigh,
		AssigneeID: targetID,
		CreatedBy:  meisterID,
		DueDate:    dueDate,
	}

	repo.On("AssignTask", ctx, tx, projectID, taskID, targetID, task.DueDate).Return(newTask, (*app_errors.AppError)(nil))

	// InsertAssignmentEvent
	repo.On("InsertAssignmentEvent", ctx, tx, mock.Anything).Return((*app_errors.AppError)(nil))

	// Commit
	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForceAufgabeHandover(ctx, meisterID, projectID, taskID, req)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, taskID, resp.AufgabenID)
	assert.Equal(t, string(entity.AufgabenInProgress), resp.Status)
	assert.Equal(t, &targetID, resp.NewAssigneeID)
	assert.Equal(t, string(entity.ActionHandoverExecute), resp.Action)
	assert.Contains(t, resp.Note, currentAssignee)
	assert.Contains(t, resp.Note, targetID)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 2: User is not a project member
func TestForceAufgabeHandover_UserNotProjectMember(t *testing.T) {
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
	targetID := "user-2"

	req := &aufgaben_dto.ForceAufgabeHandoverRequest{
		TargetID:   targetID,
		Reason:     "Better suited",
		ReasonCode: "Other",
		Note:       "Take over",
	}

	// User is not a project member
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(false, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForceAufgabeHandover(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 3: User is not a MEISTER
func TestForceAufgabeHandover_UserNotMeister(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	mitarbeiterID := "mitarbeiter-1"
	projectID := "project-1"
	taskID := "task-1"
	targetID := "user-2"

	req := &aufgaben_dto.ForceAufgabeHandoverRequest{
		TargetID:   targetID,
		Reason:     "Better suited",
		ReasonCode: "Other",
		Note:       "Take over",
	}

	// User is project member
	repo.On("CheckProjectMember", ctx, projectID, mitarbeiterID).Return(true, (*app_errors.AppError)(nil))

	currentAssignee := "user-1"
	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &currentAssignee,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// User is MITARBEITER, not MEISTER
	mitarbeiterRole := entity.MITARBEITER
	repo.On("GetUserRole", ctx, projectID, mitarbeiterID).Return(&mitarbeiterRole, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForceAufgabeHandover(ctx, mitarbeiterID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 4: Task is archived
func TestForceAufgabeHandover_TaskArchived(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	meisterID := "meister-1"
	projectID := "project-1"
	taskID := "task-1"
	targetID := "user-2"

	req := &aufgaben_dto.ForceAufgabeHandoverRequest{
		TargetID:   targetID,
		Reason:     "Better suited",
		ReasonCode: "Other",
		Note:       "Take over",
	}

	repo.On("CheckProjectMember", ctx, projectID, meisterID).Return(true, (*app_errors.AppError)(nil))

	archivedTime := time.Now().Add(-48 * time.Hour)
	currentAssignee := "user-1"
	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &currentAssignee,
		ArchivedAt:  &archivedTime, // Task is archived
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForceAufgabeHandover(ctx, meisterID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_unavailable", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 5: Task status is Done
func TestForceAufgabeHandover_TaskStatusDone(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	meisterID := "meister-1"
	projectID := "project-1"
	taskID := "task-1"
	targetID := "user-2"

	req := &aufgaben_dto.ForceAufgabeHandoverRequest{
		TargetID:   targetID,
		Reason:     "Better suited",
		ReasonCode: "Other",
		Note:       "Take over",
	}

	repo.On("CheckProjectMember", ctx, projectID, meisterID).Return(true, (*app_errors.AppError)(nil))

	currentAssignee := "user-1"
	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenDone, // Status is Done
		Priority:    entity.PriorityHigh,
		AssigneeID:  &currentAssignee,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForceAufgabeHandover(ctx, meisterID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_is_archive_or_done", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 6: Target is same as current assignee
func TestForceAufgabeHandover_TargetSameAsCurrentAssignee(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	meisterID := "meister-1"
	projectID := "project-1"
	taskID := "task-1"
	currentAssignee := "user-1"

	// Target is same as current assignee
	req := &aufgaben_dto.ForceAufgabeHandoverRequest{
		TargetID:   currentAssignee, // Same as current assignee!
		Reason:     "Better suited",
		ReasonCode: "Other",
		Note:       "Take over",
	}

	repo.On("CheckProjectMember", ctx, projectID, meisterID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &currentAssignee, // Already assigned to this user
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	meisterRole := entity.MEISTER
	repo.On("GetUserRole", ctx, projectID, meisterID).Return(&meisterRole, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForceAufgabeHandover(ctx, meisterID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.target_invalid", err.MessageKey)

	repo.AssertExpectations(t)
}

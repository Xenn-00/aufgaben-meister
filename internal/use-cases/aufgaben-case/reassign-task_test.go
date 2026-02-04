package aufgaben_case

import (
	"context"
	"testing"
	"time"

	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	worker_task "github.com/Xenn-00/aufgaben-meister/internal/worker/tasks"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test 1: Happy path - MEISTER successfully reassigns task
func TestReassignTask_Success_Meister(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	tx := new(MockTx)
	taskQueue := new(MockTaskQueue)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
		taskQueue: taskQueue,
	}

	meisterID := "meister-1"
	projectID := "project-1"
	taskID := "task-1"
	targetID := "user-2"

	reason := "Better fit for this task"
	reasonCode := "Other"
	req := &aufgaben_dto.ReassignAufgabenRequest{
		TargetID:   targetID,
		Reason:     &reason,
		ReasonCode: &reasonCode,
		Note:       "Please take over",
	}

	// Setup expectations
	// getTaskAndVerifyMember calls
	repo.On("CheckProjectMember", ctx, projectID, meisterID).Return(true, (*app_errors.AppError)(nil))

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

	// GetUserRole
	meisterRole := entity.MEISTER
	repo.On("GetUserRole", ctx, projectID, meisterID).Return(&meisterRole, (*app_errors.AppError)(nil))

	// Transaction
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// AssignTask
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
	resp, err := service.ReassignTask(ctx, meisterID, projectID, taskID, req)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, taskID, resp.AufgabenID)
	assert.Equal(t, string(entity.AufgabenInProgress), resp.Status)
	assert.Equal(t, targetID, *resp.NewAssigneeID)
	assert.Equal(t, string(entity.ActionHandoverExecute), resp.Action)
	assert.Equal(t, reason, *resp.Reason)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 2: Happy path - MITARBEITER successfully requests handover
func TestReassignTask_Success_Mitarbeiter(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	tx := new(MockTx)
	taskQueue := new(MockTaskQueue)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
		taskQueue: taskQueue,
	}

	mitarbeiterID := "mitarbeiter-1"
	projectID := "project-1"
	taskID := "task-1"
	targetID := "user-2"

	req := &aufgaben_dto.ReassignAufgabenRequest{
		TargetID: targetID,
		Note:     "Can you help me with this?",
	}

	// Setup expectations
	// getTaskAndVerifyMember calls
	repo.On("CheckProjectMember", ctx, projectID, mitarbeiterID).Return(true, (*app_errors.AppError)(nil))

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
		Priority:    entity.PriorityMedium,
		AssigneeID:  &mitarbeiterID, // Task is assigned to mitarbeiter
		DueDate:     &dueDate,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// GetUserRole
	mitarbeiterRole := entity.MITARBEITER
	repo.On("GetUserRole", ctx, projectID, mitarbeiterID).Return(&mitarbeiterRole, (*app_errors.AppError)(nil))

	// Transaction
	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	// InsertAssignmentEvent (no AssignTask for MITARBEITER, just event)
	repo.On("InsertAssignmentEvent", ctx, tx, mock.Anything).Return((*app_errors.AppError)(nil))

	// Commit
	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// EnqueueHandoverRequestNotifyMeister with proper payload type
	taskQueue.On("EnqueueHandoverRequestNotifyMeister", mock.MatchedBy(func(p *worker_task.HandoverRequestNotifyMeister) bool {
		return p.AufgabeID == taskID && p.TargetAssigneeID == targetID
	})).Return(nil)

	// Execute
	resp, err := service.ReassignTask(ctx, mitarbeiterID, projectID, taskID, req)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, taskID, resp.AufgabenID)
	assert.Equal(t, string(entity.ActionHandoverRequest), resp.Action)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
	taskQueue.AssertExpectations(t)
}

// Test 3: MEISTER missing required fields (Reason & ReasonCode)
func TestReassignTask_Meister_MissingRequiredFields(t *testing.T) {
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
	targetID := "user-2"

	// Missing Reason and ReasonCode
	req := &aufgaben_dto.ReassignAufgabenRequest{
		TargetID: targetID,
		Note:     "Please take over",
	}

	// Setup expectations
	repo.On("CheckProjectMember", ctx, projectID, meisterID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	meisterRole := entity.MEISTER
	repo.On("GetUserRole", ctx, projectID, meisterID).Return(&meisterRole, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ReassignTask(ctx, meisterID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusBadRequest, err.Code)
	assert.Equal(t, app_errors.ErrInvalidBody, err.Type)
	assert.Equal(t, "request.invalid_body", err.MessageKey)

	repo.AssertExpectations(t)
	txManager.AssertExpectations(t)
	tx.AssertExpectations(t)
}

// Test 4: MITARBEITER trying to reassign task not assigned to them
func TestReassignTask_Mitarbeiter_NotTaskAssignee(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	tx := new(MockTx)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	mitarbeiterID := "mitarbeiter-1"
	projectID := "project-1"
	taskID := "task-1"
	targetID := "user-2"
	differentUser := "different-user"

	req := &aufgaben_dto.ReassignAufgabenRequest{
		TargetID: targetID,
		Note:     "Can you help?",
	}

	// Setup expectations
	repo.On("CheckProjectMember", ctx, projectID, mitarbeiterID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityMedium,
		AssigneeID:  &differentUser, // Task assigned to different user
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	mitarbeiterRole := entity.MITARBEITER
	repo.On("GetUserRole", ctx, projectID, mitarbeiterID).Return(&mitarbeiterRole, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ReassignTask(ctx, mitarbeiterID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden.not_task_assignee", err.MessageKey)

	repo.AssertExpectations(t)
	txManager.AssertExpectations(t)
	tx.AssertExpectations(t)
}

// Test 5: User is not a project member
func TestReassignTask_UserNotProjectMember(t *testing.T) {
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

	req := &aufgaben_dto.ReassignAufgabenRequest{
		TargetID: targetID,
	}

	// User is not a project member
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(false, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ReassignTask(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

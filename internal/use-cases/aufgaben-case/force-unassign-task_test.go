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

// Test 1: Happy path - MEISTER successfully force unassigns task from target user
func TestForceUnassignTask_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	tx := new(MockTx)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "meister-1"
	projectID := "project-1"
	taskID := "task-1"
	targetID := "user-2"

	customNote := "Performance issues"
	req := &aufgaben_dto.ForceUnassignAufgabenRequest{
		TargetID:   targetID,
		Reason:     "Not meeting deadlines",
		ReasonCode: "Other",
		Note:       &customNote,
	}

	// Setup expectations
	// getTaskAndVerifyMember calls: CheckProjectMember + GetTaskByID
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &targetID, // Task is assigned to targetID
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// verifyUserRole call
	meisterRole := entity.MEISTER
	repo.On("GetUserRole", ctx, projectID, userID).Return(&meisterRole, (*app_errors.AppError)(nil))

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	unassignModel := &entity.UnassignTaskEntity{
		ID:         taskID,
		AssigneeID: targetID,
	}

	repo.On("UnassignTask", ctx, tx, mock.MatchedBy(func(m *entity.UnassignTaskEntity) bool {
		return m.AssigneeID == unassignModel.AssigneeID
	})).Return(entity.AufgabenTodo, (*app_errors.AppError)(nil))

	repo.On("InsertAssignmentEvent", ctx, tx, mock.Anything).Return((*app_errors.AppError)(nil))

	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForceUnassignTask(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, taskID, resp.AufgabenID)
	assert.Equal(t, string(entity.AufgabenTodo), resp.Status)
	assert.Equal(t, string(entity.ActionUnassign), resp.Action)
	assert.Equal(t, customNote, resp.Note)
	assert.Equal(t, "Not meeting deadlines", resp.Reason)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test 2: User is not a project member
func TestForceUnassignTask_UserNotProjectMember(t *testing.T) {
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

	req := &aufgaben_dto.ForceUnassignAufgabenRequest{
		TargetID:   targetID,
		Reason:     "Not meeting deadlines",
		ReasonCode: "Other",
	}

	// User is not a project member
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(false, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForceUnassignTask(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 3: User is not a MEISTER (doesn't have authority)
func TestForceUnassignTask_UserNotMeister(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "mitarbeiter-1"
	projectID := "project-1"
	taskID := "task-1"
	targetID := "user-2"

	req := &aufgaben_dto.ForceUnassignAufgabenRequest{
		TargetID:   targetID,
		Reason:     "Not meeting deadlines",
		ReasonCode: "Sick",
	}

	// User is project member
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &targetID,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// User has MITARBEITER role, not MEISTER
	mitarbeiterRole := entity.MITARBEITER
	repo.On("GetUserRole", ctx, projectID, userID).Return(&mitarbeiterRole, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForceUnassignTask(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 4: Task is archived
func TestForceUnassignTask_TaskArchived(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "meister-1"
	projectID := "project-1"
	taskID := "task-1"
	targetID := "user-2"

	req := &aufgaben_dto.ForceUnassignAufgabenRequest{
		TargetID:   targetID,
		Reason:     "Not meeting deadlines",
		ReasonCode: "Sick",
	}

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	archivedTime := time.Now().Add(-48 * time.Hour)
	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenArchived,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &targetID,
		ArchivedAt:  &archivedTime, // Task is archived
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForceUnassignTask(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_unavailable", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 5: Task status is Done
func TestForceUnassignTask_TaskStatusDone(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "meister-1"
	projectID := "project-1"
	taskID := "task-1"
	targetID := "user-2"

	req := &aufgaben_dto.ForceUnassignAufgabenRequest{
		TargetID:   targetID,
		Reason:     "Not meeting deadlines",
		ReasonCode: "Sick",
	}

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenDone, // Status is Done
		Priority:    entity.PriorityHigh,
		AssigneeID:  &targetID,
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForceUnassignTask(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusConflict, err.Code)
	assert.Equal(t, app_errors.ErrConflict, err.Type)
	assert.Equal(t, "conflict.task_is_archive_or_done", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test 6: Target is not the task assignee
func TestForceUnassignTask_TargetNotTaskAssignee(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	txManager := new(MockTxManager)
	service := &AufgabenService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "meister-1"
	projectID := "project-1"
	taskID := "task-1"
	targetID := "user-2"
	actualAssignee := "user-3" // Different from target

	req := &aufgaben_dto.ForceUnassignAufgabenRequest{
		TargetID:   targetID,
		Reason:     "Not meeting deadlines",
		ReasonCode: "Other",
	}

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	taskDescription := "Important task"
	task := &entity.AufgabenEntity{
		ID:          taskID,
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityHigh,
		AssigneeID:  &actualAssignee, // Task is assigned to different user
		ArchivedAt:  nil,
	}

	repo.On("GetTaskByID", ctx, taskID).Return(task, (*app_errors.AppError)(nil))

	meisterRole := entity.MEISTER
	repo.On("GetUserRole", ctx, projectID, userID).Return(&meisterRole, (*app_errors.AppError)(nil))

	// Execute
	resp, err := service.ForceUnassignTask(ctx, userID, projectID, taskID, req)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)
	assert.Equal(t, "forbidden.not_task_assignee", err.MessageKey)

	repo.AssertExpectations(t)
}

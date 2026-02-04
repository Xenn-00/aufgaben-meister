package aufgaben_case

import (
	"context"
	"fmt"

	"github.com/Xenn-00/aufgaben-meister/internal/abstraction/tx"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// verifyProjectMember checks if user is a project member
func (s *AufgabenService) verifyProjectMember(ctx context.Context, projectID, userID string) *app_errors.AppError {
	isMember, err := s.repo.CheckProjectMember(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}
	return nil
}

// validateTaskAvailability checks if task is not archived and status is valid
func (s *AufgabenService) validateTaskAvailability(task *entity.AufgabenEntity) *app_errors.AppError {
	if task.ArchivedAt != nil {
		return app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.task_unavailable", nil)
	}
	if task.Status == entity.AufgabenArchived || task.Status == entity.AufgabenDone {
		return app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.task_is_archive_or_done", fmt.Errorf("Task status is Archived or Done"))
	}
	return nil
}

// getTaskAndVerifyMember gets task and verifies user is project member
func (s *AufgabenService) getTaskAndVerifyMember(ctx context.Context, projectID, userID, taskID string) (*entity.AufgabenEntity, *app_errors.AppError) {
	if err := s.verifyProjectMember(ctx, projectID, userID); err != nil {
		return nil, err
	}

	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if err := s.validateTaskAvailability(task); err != nil {
		return nil, err
	}

	return task, nil
}

// verifyTaskAssignee checks if user is the task assignee
func (s *AufgabenService) verifyTaskAssignee(task *entity.AufgabenEntity, userID string) *app_errors.AppError {
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return app_errors.NewAppError(
			fiber.StatusForbidden,
			app_errors.ErrForbidden,
			"forbidden.not_task_assignee",
			nil,
		)
	}
	return nil
}

// verifyUserRole checks if user has the required role
func (s *AufgabenService) verifyUserRole(ctx context.Context, projectID, userID string, requiredRole entity.UserRole) *app_errors.AppError {
	userRole, err := s.repo.GetUserRole(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if userRole == nil || *userRole != requiredRole {
		return app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}
	return nil
}

// createAndInsertEvent creates a new event ID and inserts assignment event
func (s *AufgabenService) createAndInsertEvent(ctx context.Context, tx tx.Tx, event *entity.AddAssignment) (string, *app_errors.AppError) {
	eventID, idErr := uuid.NewV7()
	if idErr != nil {
		return "", app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", idErr)
	}

	event.ID = eventID.String()

	if err := s.repo.InsertAssignmentEvent(ctx, tx, event); err != nil {
		return "", err
	}

	return eventID.String(), nil
}

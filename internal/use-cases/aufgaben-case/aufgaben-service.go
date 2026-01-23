package aufgaben_case

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/dtos"
	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	aufgaben_repo "github.com/Xenn-00/aufgaben-meister/internal/repo/aufgaben-repo"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type AufgabenService struct {
	redis *redis.Client
	db    *pgxpool.Pool
	repo  aufgaben_repo.AufgabenRepoContract
}

func NewAufgabenService(db *pgxpool.Pool, redis *redis.Client) AufgabenServiceContract {
	return &AufgabenService{
		redis: redis,
		db:    db,
		repo:  aufgaben_repo.NewAufgabenRepo(db),
	}
}

func (s *AufgabenService) CreateNewAufgaben(ctx context.Context, userID, projectID string, req *aufgaben_dto.CreateNewAufgabenRequest) (*aufgaben_dto.CreateNewAufgabenResponse, *app_errors.AppError) {
	// TODO
	// Check if creator is really project member or not, doesn't care about user role
	isProjectMember, err := s.repo.CheckProjectMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	if isProjectMember == false {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}

	// Check request validity
	var assigneeID *string
	if req.AssigneeID != nil {
		member, err := s.repo.CheckProjectMember(ctx, projectID, *req.AssigneeID)
		if err != nil {
			return nil, err
		}
		if member == false {
			return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
		}
		assigneeID = req.AssigneeID
	}

	priority := entity.PriorityMedium
	if req.Priority != nil {
		priority = entity.AufgabenPriority(*req.Priority)
	}

	// Build task
	aufgabenID, idErr := uuid.NewV7()
	if idErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", idErr)
	}
	task := &entity.AufgabenEntity{
		ID:          aufgabenID.String(),
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		Status:      entity.AufgabenTodo,
		Priority:    priority,
		AssigneeID:  assigneeID,
		CreatedBy:   userID,
		DueDate:     req.DueDate,
		CreatedAt:   time.Now(),
	}

	// Insert task

	if err := s.repo.InsertNewAufgaben(ctx, task); err != nil {
		return nil, err
	}

	// Build response
	resp := &aufgaben_dto.CreateNewAufgabenResponse{
		AufgabenID:  task.ID,
		ProjectID:   projectID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		Priority:    string(task.Priority),
		AssigneeID:  task.AssigneeID,
		CreateAt:    task.CreatedAt,
		DueDate:     task.DueDate,
	}

	return resp, nil
}

func (s *AufgabenService) ListTasksProject(ctx context.Context, userID, projectID string, filter aufgaben_dto.AufgabenListFilter) ([]*aufgaben_dto.AufgabenListItem, *dtos.PaginationMeta, *app_errors.AppError) {
	// TODO
	// Check if creator is really project member or not, doesn't care about user role
	isProjectMember, err := s.repo.CheckProjectMember(ctx, projectID, userID)
	if err != nil {
		return nil, nil, err
	}

	if isProjectMember == false {
		return nil, nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}
	// Check filter validity
	if filter.AssigneeID != nil {
		member, err := s.repo.CheckProjectMember(ctx, projectID, *filter.AssigneeID)
		if err != nil {
			return nil, nil, err
		}
		if member == false {
			return nil, nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
		}
	}
	// Call repo
	tasks, err := s.repo.ListTasks(ctx, projectID, &filter)
	if err != nil {
		return nil, nil, err
	}

	totalTasks, err := s.repo.CountTasks(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}

	// Build resp
	var responses []*aufgaben_dto.AufgabenListItem
	for _, task := range tasks {
		responses = append(responses, &aufgaben_dto.AufgabenListItem{
			AufgabenID:  task.ID,
			Title:       task.Title,
			Description: task.Description,
			Status:      string(task.Status),
			Priority:    string(task.Priority),
			AssigneeID:  task.AssigneeID,
			DueDate:     task.DueDate,
		})
	}

	log.Info().Msgf("responses: %v", responses)

	totalPages := int(math.Ceil(float64(totalTasks) / float64(filter.Limit)))

	paginationMeta := &dtos.PaginationMeta{
		Page:       filter.Page,
		Limit:      filter.Limit,
		Total:      int(totalTasks),
		TotalPages: totalPages,
	}

	return responses, paginationMeta, nil
}

func (s *AufgabenService) AssignTask(ctx context.Context, userID, projectID, taskID string, req *aufgaben_dto.AufgabenAssignRequest) (*aufgaben_dto.AufgabenAssignResponse, *app_errors.AppError) {
	// TODO
	// Check if creator is really project member or not, doesn't care about user role
	isProjectMember, err := s.repo.CheckProjectMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	if isProjectMember == false {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}
	// Check if task exists
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	// Check if task is still relevant
	if task.DeletedAt != nil {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "confict.task_unavailable", nil)
	}
	// Check task status, if 'Archived' or 'Done' throw conflict
	if task.Status == entity.AufgabenArchived || task.Status == entity.AufgabenDone {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.task_is_archive_or_done", fmt.Errorf("Task status is Archived or Done"))
	}
	// Check if task is already assigned to someone
	if task.AssigneeID != nil {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.task_already_assigned_to_someone", nil)
	}
	// Check if req.due_date valid (req.due_date should > now())
	if !req.DueDate.After(time.Now()) {
		return nil, app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", fmt.Errorf("Due date must be in the future"))
	}

	// Prepare transaction to update task
	tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	assigned, err := s.repo.AssignTask(ctx, tx, projectID, taskID, userID, &req.DueDate)
	if err != nil {
		return nil, err
	}

	idEvent, idErr := uuid.NewV7()
	if idErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	note := fmt.Sprintf("First assignment by: %s", assigned.AssigneeID)
	assignEvent := &entity.AddAssignment{
		ID:         idEvent.String(),
		AufgabenID: assigned.ID,
		ActorID:    assigned.AssigneeID,
		Action:     entity.ActionAssign,
		Note:       &note,
	}

	if err := s.repo.InsertAssignmentEvent(ctx, tx, assignEvent); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	resp := &aufgaben_dto.AufgabenAssignResponse{
		AufgabenID: assigned.ID,
		ProjectID:  projectID,
		Status:     string(assigned.Status),
		Priority:   string(assigned.Priority),
		AssigneeID: assigned.AssigneeID,
		CreatedBy:  assigned.CreatedBy,
		DueDate:    assigned.DueDate,
	}

	return resp, nil
}

func (s *AufgabenService) ForwardProgressTask(ctx context.Context, userID, projectID, taskID string) (*aufgaben_dto.AufgabenForwardProgressResponse, *app_errors.AppError) {
	// TODO
	// Check if creator is really project member or not, doesn't care about user role
	isProjectMember, err := s.repo.CheckProjectMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	if isProjectMember == false {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}
	// Check if task exists
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	// Check if task is still relevant
	if task.DeletedAt != nil {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "confict.task_unavailable", nil)
	}
	// Check task status, if 'Archived' or 'Done' throw conflict
	if task.Status == entity.AufgabenArchived || task.Status == entity.AufgabenDone {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.task_is_archive_or_done", fmt.Errorf("Task status is Archived or Done"))
	}
	// Check authorithy
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return nil, app_errors.NewAppError(
			fiber.StatusForbidden,
			app_errors.ErrForbidden,
			"forbidden.not_task_assignee",
			nil,
		)
	}
	// Prepare transaction to update task
	tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	forward, err := s.repo.ForwardProgress(ctx, tx, taskID)
	if err != nil {
		return nil, err
	}

	idEvent, idErr := uuid.NewV7()
	if idErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	note := fmt.Sprintf("Assignment done by: %s, at: %s", forward.AssigneeID, time.Now().Local())
	assignEvent := &entity.AddAssignment{
		ID:         idEvent.String(),
		AufgabenID: forward.ID,
		ActorID:    forward.AssigneeID,
		Action:     entity.ActionComplete,
		Note:       &note,
	}

	if err := s.repo.InsertAssignmentEvent(ctx, tx, assignEvent); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	resp := &aufgaben_dto.AufgabenForwardProgressResponse{
		AufgabenID:  forward.ID,
		Status:      string(forward.Status),
		Priority:    string(forward.Priority),
		CreatedBy:   forward.CreatedBy,
		CompletedAt: forward.CompletedAt,
	}

	return resp, nil
}

func (s *AufgabenService) UnassignTask(ctx context.Context, userID, projectID, taskID string, req *aufgaben_dto.UnassignAufgabenRequest) (*aufgaben_dto.UnassignAufgabenResponse, *app_errors.AppError) {
	// TODO
	// Check if creator is really project member or not, doesn't care about user role
	isProjectMember, err := s.repo.CheckProjectMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	if isProjectMember == false {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}
	// Check if task exists
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	// Check if task is still relevant
	if task.DeletedAt != nil {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.task_unavailable", nil)
	}
	// Check task status, if 'Archived' or 'Done' throw conflict
	if task.Status == entity.AufgabenArchived || task.Status == entity.AufgabenDone {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.task_is_archive_or_done", fmt.Errorf("Task status is Archived or Done"))
	}
	// Check authorithy
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return nil, app_errors.NewAppError(
			fiber.StatusForbidden,
			app_errors.ErrForbidden,
			"forbidden.not_task_assignee",
			nil,
		)
	}
	// Rollback assignment status
	tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer tx.Rollback(ctx)

	idEvent, idErr := uuid.NewV7()
	if idErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	note := fmt.Sprintf("Rollback assignment status by: %v", userID)
	if req.Note != nil {
		note = *req.Note
	}

	unassignModel := &entity.UnassignTaskEntity{
		ID:         idEvent.String(),
		AssigneeID: userID,
	}

	aufgabenStatus, err := s.repo.UnassignTask(ctx, tx, unassignModel)
	if err != nil {
		return nil, err
	}

	unassignEvent := &entity.AddAssignment{
		ID:               idEvent.String(),
		AufgabenID:       taskID,
		ActorID:          userID,
		TargetAssigneeID: &userID,
		Action:           entity.ActionUnassign,
		Note:             &note,
		ReasonText:       &req.Reason,
		ReasonCode:       entity.ReasonCodeEvent(req.ReasonCode),
	}

	if err := s.repo.InsertAssignmentEvent(ctx, tx, unassignEvent); err != nil {
		return nil, err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	// Build response
	resp := &aufgaben_dto.UnassignAufgabenResponse{
		AufgabenID: taskID,
		Status:     string(aufgabenStatus),
		Note:       *unassignEvent.Note,
		Action:     string(unassignEvent.Action),
		Reason:     *unassignEvent.ReasonText,
	}

	return resp, nil
}

func (s *AufgabenService) ForceUnassignTask(ctx context.Context, userID, projectID, taskID string, req *aufgaben_dto.ForceUnassignAufgabenRequest) (*aufgaben_dto.UnassignAufgabenResponse, *app_errors.AppError) {
	// TODO
	// Check if creator is really project member or not, doesn't care about user role
	isProjectMember, err := s.repo.CheckProjectMember(ctx, projectID, req.TargetID)
	if err != nil {
		return nil, err
	}

	if isProjectMember == false {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}

	// Check if the performer has valid authority
	userRole, err := s.repo.GetUserRole(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	if userRole == nil || *userRole != entity.MEISTER {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}

	// Check if task exists
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	// Check if task is still relevant
	if task.DeletedAt != nil {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.task_unavailable", nil)
	}
	// Check task status, if 'Archived' or 'Done' throw conflict
	if task.Status == entity.AufgabenArchived || task.Status == entity.AufgabenDone {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.task_is_archive_or_done", fmt.Errorf("Task status is Archived or Done"))
	}
	// Check target validity
	if task.AssigneeID == nil || *task.AssigneeID != req.TargetID {
		return nil, app_errors.NewAppError(
			fiber.StatusForbidden,
			app_errors.ErrForbidden,
			"forbidden.not_task_assignee",
			nil,
		)
	}
	// Rollback assignment status
	tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer tx.Rollback(ctx)

	idEvent, idErr := uuid.NewV7()
	if idErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	note := fmt.Sprintf("Rollback assignment status by: %v", userID)
	if req.Note != nil {
		note = *req.Note
	}

	unassignModel := &entity.UnassignTaskEntity{
		ID:         idEvent.String(),
		AssigneeID: req.TargetID,
	}

	aufgabenStatus, err := s.repo.UnassignTask(ctx, tx, unassignModel)
	if err != nil {
		return nil, err
	}

	unassignEvent := &entity.AddAssignment{
		ID:               idEvent.String(),
		AufgabenID:       taskID,
		ActorID:          userID,
		TargetAssigneeID: &req.TargetID,
		Action:           entity.ActionUnassign,
		Note:             &note,
		ReasonText:       &req.Reason,
		ReasonCode:       entity.ReasonCodeEvent(req.ReasonCode),
	}

	if err := s.repo.InsertAssignmentEvent(ctx, tx, unassignEvent); err != nil {
		return nil, err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	// Build response
	resp := &aufgaben_dto.UnassignAufgabenResponse{
		AufgabenID: taskID,
		Status:     string(aufgabenStatus),
		Note:       *unassignEvent.Note,
		Action:     string(unassignEvent.Action),
		Reason:     *unassignEvent.ReasonText,
	}

	return resp, nil
}

func (s *AufgabenService) ReassignTask(ctx context.Context, userID, projectID, taskID string, req *aufgaben_dto.ReassignAufgabenRequest) (*aufgaben_dto.ReassignAufgabenResponse, *app_errors.AppError) {
	// TODO
	// There are 2 possible logic that can be done based who performs this endpoint
	// Meister => forcibly reassign aufgabe from one to another person that either endorsed or not.
	// Mitarbeiter => request meister to handover his task to another person that he had endorsed before.

	isProjectMember, err := s.repo.CheckProjectMember(ctx, projectID, req.TargetID)
	if err != nil {
		return nil, err
	}

	if isProjectMember == false {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}

	// Check if task exists
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	// Check if task is still relevant
	if task.DeletedAt != nil {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.task_unavailable", nil)
	}
	// Check task status, if 'Archived' or 'Done' throw conflict
	if task.Status == entity.AufgabenArchived || task.Status == entity.AufgabenDone {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.task_is_archive_or_done", fmt.Errorf("Task status is Archived or Done"))
	}

	// Check if the performer has valid authority
	userRole, err := s.repo.GetUserRole(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	// Begin Transaction
	tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer tx.Rollback(ctx)

	var resp *aufgaben_dto.ReassignAufgabenResponse
	switch *userRole {
	case entity.MEISTER:
		// Check req body
		if req.Reason == nil || req.ReasonCode == nil {
			return nil, app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", nil)
		}
		// Assign existing task to another person but without due date
		newAufgaben, err := s.repo.AssignTask(ctx, tx, projectID, taskID, req.TargetID, nil)
		if err != nil {
			return nil, err
		}
		// Insert assignment event
		note := fmt.Sprintf("Reassign this task to: %s. Because of this reason: %s", req.TargetID, *req.Reason)
		if req.Note != "" {
			note = req.Note
		}
		idEvent, idErr := uuid.NewV7()
		if idErr != nil {
			return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
		}

		reAssignmentEvent := &entity.AddAssignment{
			ID:               idEvent.String(),
			AufgabenID:       taskID,
			ActorID:          userID,
			TargetAssigneeID: &req.TargetID,
			Action:           entity.ActionHandoverExecute,
			Note:             &note,
			ReasonText:       req.Reason,
			ReasonCode:       entity.ReasonCodeEvent(*req.ReasonCode),
		}

		if err := s.repo.InsertAssignmentEvent(ctx, tx, reAssignmentEvent); err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
		}
		resp = &aufgaben_dto.ReassignAufgabenResponse{
			AufgabenID:    newAufgaben.ID,
			Status:        string(newAufgaben.Status),
			NewAssigneeID: newAufgaben.AssigneeID,
			Note:          *reAssignmentEvent.Note,
			Action:        string(reAssignmentEvent.Action),
			Reason:        *reAssignmentEvent.ReasonText,
		}

	case entity.MITARBEITER:
		// Check authorithy
		if task.AssigneeID == nil || *task.AssigneeID != userID {
			return nil, app_errors.NewAppError(
				fiber.StatusForbidden,
				app_errors.ErrForbidden,
				"forbidden.not_task_assignee",
				nil,
			)
		}
		// Insert assignment event
		note := fmt.Sprintf("Request handover this task to: %s. ", req.TargetID)
		if req.Note != "" {
			note = req.Note
		}

		idEvent, idErr := uuid.NewV7()
		if idErr != nil {
			return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
		}

		reAssignmentEvent := &entity.AddAssignment{
			ID:               idEvent.String(),
			AufgabenID:       taskID,
			ActorID:          userID,
			TargetAssigneeID: &req.TargetID,
			Action:           entity.ActionHandoverRequest,
			Note:             &note,
		}

		if err := s.repo.InsertAssignmentEvent(ctx, tx, reAssignmentEvent); err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
		}
		resp = &aufgaben_dto.ReassignAufgabenResponse{
			AufgabenID: taskID,
			Note:       *reAssignmentEvent.Note,
			Action:     string(reAssignmentEvent.Action),
			Reason:     *reAssignmentEvent.ReasonText,
		}
	}

	return resp, nil
}

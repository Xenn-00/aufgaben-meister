package aufgaben_case

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/abstraction/cache"
	"github.com/Xenn-00/aufgaben-meister/internal/abstraction/tx"
	"github.com/Xenn-00/aufgaben-meister/internal/dtos"
	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/Xenn-00/aufgaben-meister/internal/queue"
	aufgaben_repo "github.com/Xenn-00/aufgaben-meister/internal/repo/aufgaben-repo"
	worker_task "github.com/Xenn-00/aufgaben-meister/internal/worker/tasks"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type AufgabenService struct {
	cache     cache.Cache
	txManager tx.TxManager
	repo      aufgaben_repo.AufgabenRepoContract
	taskQueue queue.TaskQueueClient
}

func NewAufgabenService(db *pgxpool.Pool, redis *redis.Client) AufgabenServiceContract {
	return &AufgabenService{
		cache:     cache.NewRedisCache(redis),
		txManager: tx.NewPgxTxManager(db),
		repo:      aufgaben_repo.NewAufgabenRepo(db),
		taskQueue: queue.NewTaskQueue(redis),
	}
}

func (s *AufgabenService) CreateNewAufgaben(ctx context.Context, userID, projectID string, req *aufgaben_dto.CreateNewAufgabenRequest) (*aufgaben_dto.CreateNewAufgabenResponse, *app_errors.AppError) {
	// TODO
	// Check if creator is really project member or not, doesn't care about user role
	if err := s.verifyProjectMember(ctx, projectID, userID); err != nil {
		return nil, err
	}

	// Check request validity
	var assigneeID *string
	if req.AssigneeID != nil {
		if err := s.verifyProjectMember(ctx, projectID, *req.AssigneeID); err != nil {
			return nil, err
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

func (s *AufgabenService) ListTasksProject(ctx context.Context, userID, projectID string, filter aufgaben_dto.AufgabenListFilter) ([]*aufgaben_dto.AufgabenItem, *dtos.PaginationMeta, *app_errors.AppError) {
	// TODO
	// Check if user is really project member or not, doesn't care about user role
	if err := s.verifyProjectMember(ctx, projectID, userID); err != nil {
		return nil, nil, err
	}

	// Check filter validity
	if filter.AssigneeID != nil {
		if err := s.verifyProjectMember(ctx, projectID, *filter.AssigneeID); err != nil {
			return nil, nil, err
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
	var responses []*aufgaben_dto.AufgabenItem
	for _, task := range tasks {
		responses = append(responses, &aufgaben_dto.AufgabenItem{
			AufgabenID:  task.ID,
			Title:       task.Title,
			Description: task.Description,
			Status:      string(task.Status),
			Priority:    string(task.Priority),
			AssigneeID:  task.AssigneeID,
			DueDate:     task.DueDate,
		})
	}

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
	// Check if user is really project member or not, doesn't care about user role, returning task
	task, err := s.getTaskAndVerifyMember(ctx, projectID, userID, taskID)
	if err != nil {
		return nil, err
	}
	// Check if task is still relevant
	if err := s.validateTaskAvailability(task); err != nil {
		return nil, err
	}
	// Check if task is already assigned to someone
	if task.AssigneeID != nil {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.task_already_assigned_to_someone", nil)
	}
	// Check if req.due_date valid (req.due_date should > now() + 1 hour)
	if !req.DueDate.After(time.Now().Add(1 * time.Hour)) {
		return nil, app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", fmt.Errorf("Due date must be in the future"))
	}

	// Prepare transaction to update task
	tx, txErr := s.txManager.Begin(ctx)
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	assigned, err := s.repo.AssignTask(ctx, tx, projectID, taskID, userID, &req.DueDate)
	if err != nil {
		return nil, err
	}

	note := fmt.Sprintf("First assignment by: %s", assigned.AssigneeID)
	assignEvent := &entity.AddAssignment{
		AufgabenID: assigned.ID,
		ActorID:    assigned.AssigneeID,
		Action:     entity.ActionAssign,
		Note:       &note,
	}

	if _, err := s.createAndInsertEvent(ctx, tx, assignEvent); err != nil {
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

func (s *AufgabenService) GetAufgabeDetails(ctx context.Context, userID, projectID, taskID string) (*aufgaben_dto.AufgabenItem, *app_errors.AppError) {
	// TODO
	// Check if performer is really project member or not, doesn't care about user role
	if err := s.verifyProjectMember(ctx, projectID, userID); err != nil {
		return nil, err
	}

	// Check cache
	cacheKey := fmt.Sprintf("aufgaben:details:%s", taskID)
	cacheData, cacheErr := s.cache.Get(ctx, cacheKey)
	if cacheData != nil && cacheErr == nil {
		// cache.Get returns *any, so dereference and type assert to the expected type
		if cached, ok := (*cacheData).(*aufgaben_dto.AufgabenItem); ok && cached != nil {
			return cached, nil
		}
		// If cached data has unexpected type, log and continue to fetch from DB
		log.Warn().Msgf("unexpected cache type for key %s", cacheKey)
	}

	// Cache miss, fetch from DB

	// Check if task exists
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Build resp
	resp := &aufgaben_dto.AufgabenItem{
		AufgabenID:  task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		Priority:    string(task.Priority),
		AssigneeID:  task.AssigneeID,
		DueDate:     task.DueDate,
	}

	// cache task details in redis
	if err := s.cache.Set(ctx, cacheKey, resp, 5*time.Minute); err != nil {
		log.Error().Err(err.Err).Msg("Fehler beim Setzen des Redis-Cache")
		return nil, err
	}

	return resp, nil
}

func (s *AufgabenService) ForwardProgressTask(ctx context.Context, userID, projectID, taskID string) (*aufgaben_dto.AufgabenForwardProgressResponse, *app_errors.AppError) {
	// TODO
	// Check if user is really project member or not, doesn't care about user role, returning task
	task, err := s.getTaskAndVerifyMember(ctx, projectID, userID, taskID)
	if err != nil {
		return nil, err
	}

	// Check if task is still relevant
	if err := s.validateTaskAvailability(task); err != nil {
		return nil, err
	}

	// Check authorithy
	if err := s.verifyTaskAssignee(task, userID); err != nil {
		return nil, err
	}

	// Prepare transaction to update task
	tx, txErr := s.txManager.Begin(ctx)
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	forward, err := s.repo.ForwardProgress(ctx, tx, taskID)
	if err != nil {
		return nil, err
	}

	note := fmt.Sprintf("Assignment done by: %s, at: %s", forward.AssigneeID, time.Now().Local())
	completeEvent := &entity.AddAssignment{
		AufgabenID: forward.ID,
		ActorID:    forward.AssigneeID,
		Action:     entity.ActionComplete,
		Note:       &note,
	}

	if _, err := s.createAndInsertEvent(ctx, tx, completeEvent); err != nil {
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
	// Check if creator is really project member or not and returning corresponding task, doesn't care about user role
	task, err := s.getTaskAndVerifyMember(ctx, projectID, userID, taskID)
	if err != nil {
		return nil, err
	}
	// Check if task is still relevant
	if err := s.validateTaskAvailability(task); err != nil {
		return nil, err
	}
	// Check authorithy
	if err := s.verifyTaskAssignee(task, userID); err != nil {
		return nil, err
	}

	// Rollback assignment status
	tx, txErr := s.txManager.Begin(ctx)
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer tx.Rollback(ctx)

	note := fmt.Sprintf("Rollback assignment status by: %v", userID)
	if req.Note != nil {
		note = *req.Note
	}

	unassignModel := &entity.UnassignTaskEntity{
		ID:         task.ID,
		AssigneeID: userID,
	}

	aufgabenStatus, err := s.repo.UnassignTask(ctx, tx, unassignModel)
	if err != nil {
		return nil, err
	}

	unassignEvent := &entity.AddAssignment{
		AufgabenID:       taskID,
		ActorID:          userID,
		TargetAssigneeID: &userID,
		Action:           entity.ActionUnassign,
		Note:             &note,
		ReasonText:       &req.Reason,
		ReasonCode:       entity.ReasonCodeEvent(req.ReasonCode),
	}

	if _, err := s.createAndInsertEvent(ctx, tx, unassignEvent); err != nil {
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
	// Check if user is really project member or not, doesn't care about user role, returning task
	task, err := s.getTaskAndVerifyMember(ctx, projectID, userID, taskID)
	if err != nil {
		return nil, err
	}

	// Check if the performer has valid authority
	if err := s.verifyUserRole(ctx, projectID, userID, entity.MEISTER); err != nil {
		return nil, err
	}

	// Check if task is still relevant
	if err := s.validateTaskAvailability(task); err != nil {
		return nil, err
	}

	// Check target validity
	if err := s.verifyTaskAssignee(task, req.TargetID); err != nil {
		return nil, err
	}
	// Rollback assignment status
	tx, txErr := s.txManager.Begin(ctx)
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", txErr)
	}
	defer tx.Rollback(ctx)

	note := fmt.Sprintf("Rollback assignment status by: %v", userID)
	if req.Note != nil {
		note = *req.Note
	}

	unassignModel := &entity.UnassignTaskEntity{
		ID:         task.ID,
		AssigneeID: req.TargetID,
	}

	aufgabenStatus, err := s.repo.UnassignTask(ctx, tx, unassignModel)
	if err != nil {
		return nil, err
	}

	unassignEvent := &entity.AddAssignment{
		AufgabenID:       taskID,
		ActorID:          userID,
		TargetAssigneeID: &req.TargetID,
		Action:           entity.ActionUnassign,
		Note:             &note,
		ReasonText:       &req.Reason,
		ReasonCode:       entity.ReasonCodeEvent(req.ReasonCode),
	}

	if _, err := s.createAndInsertEvent(ctx, tx, unassignEvent); err != nil {
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

	task, err := s.getTaskAndVerifyMember(ctx, projectID, userID, taskID)
	if err != nil {
		return nil, err
	}

	// Check if task is still relevant
	if err := s.validateTaskAvailability(task); err != nil {
		return nil, err
	}

	// Check if the performer has valid authority
	userRole, err := s.repo.GetUserRole(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	// Begin Transaction
	tx, txErr := s.txManager.Begin(ctx)
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
		newAufgaben, err := s.repo.AssignTask(ctx, tx, projectID, taskID, req.TargetID, task.DueDate)
		if err != nil {
			return nil, err
		}
		// Insert assignment event
		note := fmt.Sprintf("Reassign this task to: %s. Because of this reason: %s", req.TargetID, *req.Reason)
		if req.Note != "" {
			note = req.Note
		}

		reAssignmentEvent := &entity.AddAssignment{
			AufgabenID:       taskID,
			ActorID:          userID,
			TargetAssigneeID: &req.TargetID,
			Action:           entity.ActionHandoverExecute,
			Note:             &note,
			ReasonText:       req.Reason,
			ReasonCode:       entity.ReasonCodeEvent(*req.ReasonCode),
		}

		if _, err := s.createAndInsertEvent(ctx, tx, reAssignmentEvent); err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
		}
		resp = &aufgaben_dto.ReassignAufgabenResponse{
			AufgabenID:    newAufgaben.ID,
			Status:        string(newAufgaben.Status),
			NewAssigneeID: &newAufgaben.AssigneeID,
			Note:          *reAssignmentEvent.Note,
			Action:        string(reAssignmentEvent.Action),
			Reason:        reAssignmentEvent.ReasonText,
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

		reAssignmentEvent := &entity.AddAssignment{
			AufgabenID:       taskID,
			ActorID:          userID,
			TargetAssigneeID: &req.TargetID,
			Action:           entity.ActionHandoverRequest,
			Note:             &note,
		}

		if _, err := s.createAndInsertEvent(ctx, tx, reAssignmentEvent); err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
		}

		// Enqueue this task so that meister can be reminded
		payloadTask := &worker_task.HandoverRequestNotifyMeister{
			AufgabeID:        taskID,
			ProjectName:      *task.ProjectName,
			ProjectID:        projectID,
			AufgabeTitle:     task.Title,
			AufgabeStatus:    string(task.Status),
			AssigneeID:       userID,
			TargetAssigneeID: req.TargetID,
			RequestedAt:      time.Now(),
			DueDate:          *task.DueDate,
			Note:             note,
		}
		if err := s.taskQueue.EnqueueHandoverRequestNotifyMeister(payloadTask); err != nil {
			log.Error().Err(err).Msg("Fehler beim Stellen die Aufgabe in die Warteschlange")
		}

		resp = &aufgaben_dto.ReassignAufgabenResponse{
			AufgabenID: taskID,
			Note:       *reAssignmentEvent.Note,
			Status:     string(task.Status),
			Action:     string(reAssignmentEvent.Action),
		}
	}

	return resp, nil
}

func (s *AufgabenService) ListAssignedTasks(ctx context.Context, userID string, filter *aufgaben_dto.AssignedAufgabenFilter) ([]*aufgaben_dto.AssignedAufgabenListItem, *dtos.CursorPaginationMeta, *app_errors.AppError) {
	// TODO
	// Check filter validity
	// 1. check limit
	if filter.Limit == 0 {
		filter.Limit = 20
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}
	// 2. check project_id if provided
	if filter.ProjectID != nil {
		isProjectMember, err := s.repo.CheckProjectMember(ctx, *filter.ProjectID, userID)
		if err != nil {
			return nil, nil, err
		}
		if isProjectMember == false {
			return nil, nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
		}
	}
	// 3. check cursor validity if provided
	if filter.Cursor != nil {
		_, err := uuid.Parse(*filter.Cursor)
		if err != nil {
			return nil, nil, app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidQuery, "request.invalid_query", nil)
		}
	}
	// Call repo
	tasks, err := s.repo.ListAssignedTasks(ctx, userID, filter)
	if err != nil {
		return nil, nil, err
	}

	// Build response cursor
	hasMore := false
	if len(tasks) > filter.Limit {
		hasMore = true
		tasks = tasks[:filter.Limit]
	}

	var nextCursor *string
	if hasMore {
		nextCursor = &tasks[len(tasks)-1].ID
	}

	var data []*aufgaben_dto.AssignedAufgabenListItem
	for _, task := range tasks {
		data = append(data, &aufgaben_dto.AssignedAufgabenListItem{
			AufgabenID:  task.ID,
			ProjectName: task.ProjectName,
			Title:       task.Title,
			Description: task.Description,
			Status:      string(task.Status),
			Priority:    string(task.Priority),
			DueDate:     task.DueDate,
		})
	}

	cursor := &dtos.CursorPaginationMeta{
		Limit:      filter.Limit,
		NextCursor: *nextCursor,
		HasMore:    hasMore,
	}

	return data, cursor, nil
}

func (s *AufgabenService) ArchiveTask(ctx context.Context, userID, projectID, taskID string) *app_errors.AppError {
	// TODO
	// Check if creator is really project member or not
	task, err := s.getTaskAndVerifyMember(ctx, projectID, userID, taskID)
	if err != nil {
		return err
	}

	// Check if task is still relevant
	if err := s.validateTaskAvailability(task); err != nil {
		return err
	}

	// Check if the performer has valid authority
	if err := s.verifyUserRole(ctx, projectID, userID, entity.MEISTER); err != nil {
		return err
	}

	// Archive the task, start transaction
	tx, txErr := s.txManager.Begin(ctx)
	if txErr != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.ArchiveTask(ctx, tx, taskID); err != nil {
		return err
	}

	note := fmt.Sprintf("Archive task by: %s", userID)
	archivedAt := time.Now()
	archiveEvent := &entity.AddAssignment{
		AufgabenID:     taskID,
		ActorID:        userID,
		Action:         entity.ActionArchive,
		Note:           &note,
		TaskArchivedAt: &archivedAt,
		ArchivedBy:     &userID,
	}

	// insert archive event
	if _, err := s.createAndInsertEvent(ctx, tx, archiveEvent); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", nil)
	}

	return nil
}

func (s *AufgabenService) UpdateDueDate(ctx context.Context, userID, projectID, taskID string, req *aufgaben_dto.UpdateDueDateRequest) (*aufgaben_dto.UpdateDueDateResponse, *app_errors.AppError) {
	// TODO
	// Check if performer is really project member or not
	task, err := s.getTaskAndVerifyMember(ctx, projectID, userID, taskID)
	if err != nil {
		return nil, err
	}

	// Check if task is still relevant
	if err := s.validateTaskAvailability(task); err != nil {
		return nil, err
	}

	// Update due date
	tx, txErr := s.txManager.Begin(ctx)
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer tx.Rollback(ctx)

	updatedDueDate, err := s.repo.UpdateDueDate(ctx, tx, taskID, req.DueDate)
	if err != nil {
		return nil, err
	}

	note := fmt.Sprintf("Task due date updated to: %s", req.DueDate)
	updateEvent := &entity.AddAssignment{
		AufgabenID: taskID,
		ActorID:    userID,
		Action:     entity.ActionDueDateUpdate,
		Note:       &note,
	}

	if _, err := s.createAndInsertEvent(ctx, tx, updateEvent); err != nil {
		return nil, err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	// Build response
	resp := &aufgaben_dto.UpdateDueDateResponse{
		AufgabenID: taskID,
		DueDate:    *updatedDueDate,
	}

	return resp, nil
}

func (s *AufgabenService) FetchEventsForTask(ctx context.Context, userID, projectID, taskID string, filters *aufgaben_dto.AufgabenEventFilter) ([]*aufgaben_dto.AufgabenEventItem, *dtos.CursorPaginationMeta, *app_errors.AppError) {
	// TODO
	// Check if performer is really project member or not
	if err := s.verifyProjectMember(ctx, projectID, userID); err != nil {
		return nil, nil, err
	}

	// Check if task exists
	if _, err := s.repo.GetTaskByID(ctx, taskID); err != nil {
		return nil, nil, err
	}

	// We want to allow fetching events even the task is archived or done
	// We need to verify filters first
	if filters.Limit == 0 {
		filters.Limit = 20
	} else if filters.Limit > 100 {
		filters.Limit = 100
	}

	if filters.Cursor != nil {
		_, err := uuid.Parse(*filters.Cursor)
		if err != nil {
			return nil, nil, app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidQuery, "request.invalid_query", nil)
		}
	}

	// Call repo
	events, err := s.repo.ListEventsForTask(ctx, taskID, filters)
	if err != nil {
		return nil, nil, err
	}

	// Build response cursor
	hasMore := false
	if len(events) > filters.Limit {
		hasMore = true
		events = events[:filters.Limit]
	}

	var nextCursor *string
	if hasMore {
		nextCursor = &events[len(events)-1].ID
	}

	var data []*aufgaben_dto.AufgabenEventItem
	for _, event := range events {
		data = append(data, &aufgaben_dto.AufgabenEventItem{
			EventID:     event.ID,
			AufgabeID:   event.AufgabenID,
			ActorID:     event.ActorID,
			EventAction: string(event.Action),
			Note:        event.Note,
			TargetID:    event.TargetAssigneeID,
			ReasonCode:  (*string)(event.ReasonCode),
			ReasonText:  event.ReasonText,
			EventTime:   event.CreatedAt,
		})
	}

	return data, &dtos.CursorPaginationMeta{
		Limit:      filters.Limit,
		NextCursor: *nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *AufgabenService) ForceAufgabeHandover(ctx context.Context, userID, projectID, taskID string, req *aufgaben_dto.ForceAufgabeHandoverRequest) (*aufgaben_dto.ReassignAufgabenResponse, *app_errors.AppError) {
	// TODO
	// Check if creator is really project member or not, doesn't care about user role
	task, err := s.getTaskAndVerifyMember(ctx, projectID, userID, taskID)
	if err != nil {
		return nil, err
	}

	// Check if the performer has valid authority
	if err := s.verifyUserRole(ctx, projectID, userID, entity.MEISTER); err != nil {
		return nil, err
	}

	// Check if task is still relevant
	if err := s.validateTaskAvailability(task); err != nil {
		return nil, err
	}

	// Check target validity, target should be different from current assignee
	if *task.AssigneeID == req.TargetID {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict.target_invalid", fmt.Errorf("Target and current assignee can't be the same person"))
	}

	// Handover assignment
	tx, txErr := s.txManager.Begin(ctx)
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", txErr)
	}
	defer tx.Rollback(ctx)

	newAufgabe, err := s.repo.AssignTask(ctx, tx, projectID, taskID, req.TargetID, task.DueDate)
	if err != nil {
		return nil, err
	}

	// Insert handover event

	note := fmt.Sprintf("Force handover assignment from %s to %s, note: %s", *task.AssigneeID, req.TargetID, req.Note)

	handoverEvent := &entity.AddAssignment{
		AufgabenID:       taskID,
		ActorID:          userID,
		TargetAssigneeID: &req.TargetID,
		Action:           entity.ActionHandoverExecute,
		Note:             &note,
		ReasonText:       &req.Reason,
		ReasonCode:       entity.ReasonCodeEvent(req.ReasonCode),
	}

	if _, err := s.createAndInsertEvent(ctx, tx, handoverEvent); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	resp := &aufgaben_dto.ReassignAufgabenResponse{
		AufgabenID:    newAufgabe.ID,
		Status:        string(newAufgabe.Status),
		NewAssigneeID: &newAufgabe.AssigneeID,
		Note:          *handoverEvent.Note,
		Action:        string(handoverEvent.Action),
		Reason:        handoverEvent.ReasonText,
	}

	return resp, nil
}

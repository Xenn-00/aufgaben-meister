package aufgaben_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AufgabenRepo struct {
	db *pgxpool.Pool
}

func NewAufgabenRepo(db *pgxpool.Pool) AufgabenRepoContract {
	return &AufgabenRepo{
		db: db,
	}
}

func (r *AufgabenRepo) CheckProjectMember(ctx context.Context, projectID, userID string) (bool, *app_errors.AppError) {
	query := `
	SELECT EXISTS (
		SELECT 1
		FROM project_members
		WHERE project_id = $1
			AND user_id = $2 
	);
	`

	var exists bool
	if err := r.db.QueryRow(ctx, query, projectID, userID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
		}
		return false, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	return exists, nil
}

func (r *AufgabenRepo) GetUserRole(ctx context.Context, projectID, userID string) (*entity.UserRole, *app_errors.AppError) {
	query := `
	SELECT role FROM project_members
	WHERE project_id = $1
		AND user_id = $2;
	`

	var role entity.UserRole
	if err := r.db.QueryRow(ctx, query, projectID, userID).Scan(&role); err != nil {
		return nil, app_errors.MapPgxError(err)
	}
	return &role, nil
}

func (r *AufgabenRepo) GetTaskByID(ctx context.Context, taskID string) (*entity.AufgabenEntity, *app_errors.AppError) {
	query := `
	SELECT * FROM aufgaben
	WHERE id = $1;
	`

	var row entity.AufgabenEntity
	if err := r.db.QueryRow(ctx, query, taskID).Scan(&row.ID, &row.ProjectID, &row.Title, &row.Description, &row.Status, &row.Priority, &row.AssigneeID, &row.CreatedBy, &row.DueDate, &row.CreatedAt, &row.UpdatedAt, &row.DeletedAt, &row.CompletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "task_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return &row, nil
}

func (r *AufgabenRepo) InsertNewAufgaben(ctx context.Context, task *entity.AufgabenEntity) *app_errors.AppError {
	query := `
	INSERT INTO aufgaben (
			id,
			project_id,
			title,
			description,
			status,
			priority,
			assignee_id,
			created_by,
			due_date,
			created_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10
		)
	`

	if _, err := r.db.Exec(
		ctx,
		query,
		task.ID,
		task.ProjectID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.AssigneeID,
		task.CreatedBy,
		task.DueDate,
		task.CreatedAt,
	); err != nil {
		return app_errors.MapPgxError(err)
	}

	return nil
}

func (r *AufgabenRepo) CountTasks(ctx context.Context, projectID string) (int64, *app_errors.AppError) {
	query := `
	SELECT COUNT(*)
	FROM aufgaben
	WHERE project_id = $1
		AND status != 'Archived'
		AND deleted_at = NULL;
	`
	var count int64
	if err := r.db.QueryRow(ctx, query, projectID).Scan(&count); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
		}
		return 0, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return count, nil
}

func (r *AufgabenRepo) ListTasks(ctx context.Context, projectID string, filter *aufgaben_dto.AufgabenListFilter) ([]entity.AufgabenEntity, *app_errors.AppError) {
	query := `
	SELECT * FROM aufgaben
	WHERE project_id = $1
		AND deleted_at IS NULL
	`

	args := []any{projectID}
	argsPos := 2

	if filter.Status != nil {
		if *filter.Status != string(entity.AufgabenArchived) {
			query += fmt.Sprintf(" AND status = $%d", argsPos)
			args = append(args, filter.Status)
			argsPos++
		}
	}

	if filter.AssigneeID != nil {
		query += fmt.Sprintf(" AND assignee_id = $%d", argsPos)
		args = append(args, filter.AssigneeID)
		argsPos++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d;", argsPos, argsPos+1)

	offset := (filter.Page - 1) * filter.Limit

	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	var results []entity.AufgabenEntity
	for rows.Next() {
		var result entity.AufgabenEntity
		if err := rows.Scan(&result.ID, &result.ProjectID, &result.Title, &result.Description, &result.Status, &result.Priority, &result.AssigneeID, &result.CreatedBy, &result.DueDate, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt, &result.CompletedAt); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
			}
			return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, app_errors.MapPgxError(err)
	}

	return results, nil
}

func (r *AufgabenRepo) AssignTask(ctx context.Context, tx pgx.Tx, projectID, taskID, userID string, dueDate *time.Time) (*entity.AssignTaskEntity, *app_errors.AppError) {
	query := `
	UPDATE aufgaben
	SET status = 'In_Progress',
		assignee_id = $1,
		due_date = $2,
		updated_at = now()
	WHERE project_id = $3
		AND id = $4
	RETURNING id, status, priority, assignee_id, created_by, due_date;
	`

	var rows entity.AssignTaskEntity
	if err := tx.QueryRow(ctx, query, userID, dueDate, projectID, taskID).Scan(&rows.ID, &rows.Status, &rows.Priority, &rows.AssigneeID, &rows.CreatedBy, &rows.DueDate); err != nil {
		return nil, app_errors.MapPgxError(err)
	}

	return &rows, nil
}

func (r *AufgabenRepo) ForwardProgress(ctx context.Context, tx pgx.Tx, taskID string) (*entity.CompleteTaskEntity, *app_errors.AppError) {
	query := `
	UPDATE aufgaben
	SET status = 'Done',
		completed_at = now()
	WHERE id = $1;
	RETURNING id, status, priority, assignee_id, created_by, complete_at;
	`

	var rows entity.CompleteTaskEntity
	if err := tx.QueryRow(ctx, query, taskID).Scan(&rows.ID, &rows.Status, &rows.Priority, &rows.AssigneeID, &rows.CreatedBy, &rows.CompletedAt); err != nil {
		return nil, app_errors.MapPgxError(err)
	}

	return &rows, nil
}

func (r *AufgabenRepo) InsertAssignmentEvent(ctx context.Context, tx pgx.Tx, event *entity.AddAssignment) *app_errors.AppError {
	query := `
	INSERT INTO aufgaben_assignment_events (
		id,
		aufgaben_id,
		actor_id,
		target_assignee_id,
		action,
		note,
		reason_code,
		reason_text
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8
	);
	`
	if _, err := tx.Exec(ctx, query, event.ID, event.AufgabenID, event.ActorID, event.TargetAssigneeID, event.Action, event.Note, event.ReasonCode, event.ReasonText); err != nil {
		return app_errors.MapPgxError(err)
	}
	return nil
}

func (r *AufgabenRepo) UnassignTask(ctx context.Context, tx pgx.Tx, rollbackModel *entity.UnassignTaskEntity) (entity.AufgabenStatus, *app_errors.AppError) {
	// 1. Update aufgaben state
	queryAufgaben := `
	UPDATE aufgaben
	SET status = 'Todo',
		assignee_id = NULL,
		due_date = NULL
	WHERE id = $1
		AND assignee_id = $2
	RETURNING status;
	`
	var status entity.AufgabenStatus
	if err := tx.QueryRow(ctx, queryAufgaben, rollbackModel.ID, rollbackModel.AssigneeID).Scan(&status); err != nil {
		return "", app_errors.MapPgxError(err)
	}

	return status, nil
}

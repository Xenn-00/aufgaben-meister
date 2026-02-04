package aufgaben_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/abstraction/tx"
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
	SELECT a.*, p.name FROM aufgaben a
	JOIN projects p ON p.id = a.project_id
	WHERE id = $1;
	`

	var row entity.AufgabenEntity
	if err := r.db.QueryRow(ctx, query, taskID).Scan(&row.ID, &row.ProjectID, &row.Title, &row.Description, &row.Status, &row.Priority, &row.AssigneeID, &row.CreatedBy, &row.DueDate, &row.CreatedAt, &row.UpdatedAt, &row.ArchivedAt, &row.CompletedAt, &row.ProjectName); err != nil {
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
		AND archived_at = NULL;
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
	SELECT a.*, p.name FROM aufgaben a
	JOIN projects p ON p.id = a.project_id
	WHERE project_id = $1
		AND archived_at IS NULL
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
		if err := rows.Scan(&result.ID, &result.ProjectID, &result.Title, &result.Description, &result.Status, &result.Priority, &result.AssigneeID, &result.CreatedBy, &result.DueDate, &result.CreatedAt, &result.UpdatedAt, &result.ArchivedAt, &result.CompletedAt, &result.ProjectName); err != nil {
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

func (r *AufgabenRepo) AssignTask(ctx context.Context, t tx.Tx, projectID, taskID, userID string, dueDate *time.Time) (*entity.AssignTaskEntity, *app_errors.AppError) {
	pgxTx := t.(*tx.PgxTx).Tx
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
	if err := pgxTx.QueryRow(ctx, query, userID, dueDate, projectID, taskID).Scan(&rows.ID, &rows.Status, &rows.Priority, &rows.AssigneeID, &rows.CreatedBy, &rows.DueDate); err != nil {
		return nil, app_errors.MapPgxError(err)
	}

	return &rows, nil
}

func (r *AufgabenRepo) ForwardProgress(ctx context.Context, t tx.Tx, taskID string) (*entity.CompleteTaskEntity, *app_errors.AppError) {
	pgxTx := t.(*tx.PgxTx).Tx

	query := `
	UPDATE aufgaben
	SET status = 'Done',
		completed_at = now()
	WHERE id = $1;
	RETURNING id, status, priority, assignee_id, created_by, complete_at;
	`

	var rows entity.CompleteTaskEntity
	if err := pgxTx.QueryRow(ctx, query, taskID).Scan(&rows.ID, &rows.Status, &rows.Priority, &rows.AssigneeID, &rows.CreatedBy, &rows.CompletedAt); err != nil {
		return nil, app_errors.MapPgxError(err)
	}

	return &rows, nil
}

func (r *AufgabenRepo) InsertAssignmentEvent(ctx context.Context, t tx.Tx, event *entity.AddAssignment) *app_errors.AppError {
	pgxTx := t.(*tx.PgxTx).Tx
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
	if _, err := pgxTx.Exec(ctx, query, event.ID, event.AufgabenID, event.ActorID, event.TargetAssigneeID, event.Action, event.Note, event.ReasonCode, event.ReasonText); err != nil {
		return app_errors.MapPgxError(err)
	}
	return nil
}

func (r *AufgabenRepo) UnassignTask(ctx context.Context, t tx.Tx, rollbackModel *entity.UnassignTaskEntity) (entity.AufgabenStatus, *app_errors.AppError) {
	pgxTx := t.(*tx.PgxTx).Tx
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
	if err := pgxTx.QueryRow(ctx, queryAufgaben, rollbackModel.ID, rollbackModel.AssigneeID).Scan(&status); err != nil {
		return "", app_errors.MapPgxError(err)
	}

	return status, nil
}

func (r *AufgabenRepo) ShouldRemind(ctx context.Context, taskID string) (*entity.ReminderAufgaben, *app_errors.AppError) {
	query := `
	SELECT a.id, a.project_id, a.title, a.status, a.priority, a.assignee_id,
	a.due_date, a.last_reminder_at, u.email as assignee_email, p.name as project_name
	FROM aufgaben a
	JOIN users u ON u.id = a.assignee_id
	JOIN projects p ON p.id = a.project_id
	WHERE a.id = $1
		AND a.status = 'In_Progress'
		AND a.assignee_id IS NOT NULL
		AND a.due_date IS NOT NULL
		AND a.reminder_stage = 'None'
		AND a.last_reminder_at IS NULL
		AND now() >= a.due_date - INTERVAL '1 hour';
	`

	var row entity.ReminderAufgaben
	if err := r.db.QueryRow(ctx, query, taskID).Scan(&row.ID, &row.ProjectID, &row.Title, &row.Status, &row.Priority, &row.AssigneeID, &row.DueDate, &row.LastReminderAt, &row.EmailAssignee, &row.ProjectName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return &row, nil
}

func (r *AufgabenRepo) ListShouldRemindOverdue(ctx context.Context) ([]entity.ReminderAufgaben, *app_errors.AppError) {
	query := `
	SELECT a.id, a.project_id, a.title, a.status, a.priority, a.assignee_id,
	a.due_date, a.last_reminder_at, u.email as assignee_email, p.name as project_name
	FROM aufgaben a
	JOIN users u ON u.id = a.assignee_id
	JOIN projects p ON p.id = a.project_id
	WHERE a.status = 'In_Progress'
		AND a.assignee_id IS NOT NULL
		AND a.due_date IS NOT NULL
		AND a.reminder_stage = 'Before_Due'
		AND (
			last_reminder_at IS NULL
			OR now() >= last_reminder_at + INTERVAL '24 hours'
		)
		AND now() >= a.due_date;
	`

	var aufgaben []entity.ReminderAufgaben
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, app_errors.MapPgxError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var aufgabe entity.ReminderAufgaben
		if err := rows.Scan(&aufgabe.ID, &aufgabe.ProjectID, &aufgabe.Title, &aufgabe.Status, &aufgabe.Priority, &aufgabe.AssigneeID, &aufgabe.DueDate, &aufgabe.LastReminderAt, &aufgabe.EmailAssignee, &aufgabe.ProjectName); err != nil {
			return nil, app_errors.MapPgxError(err)
		}
		aufgaben = append(aufgaben, aufgabe)
	}

	return aufgaben, nil
}

func (r *AufgabenRepo) BatchUpdateAufgabenReminderOverdue(ctx context.Context, t tx.Tx, taskIDs []string) *app_errors.AppError {
	pgxTx := t.(*tx.PgxTx).Tx
	query := `
	UPDATE aufgaben
	SET reminder_stage = 'Overdue',
		last_reminder_at = now()
	WHERE id = $1 AND reminder_stage = 'Before_Due';
	`

	batch := &pgx.Batch{}

	for _, taskID := range taskIDs {
		batch.Queue(query, taskID)
	}

	br := pgxTx.SendBatch(ctx, batch)

	err := br.Close()
	if err != nil {
		return app_errors.MapPgxError(err)
	}

	return nil
}

func (r *AufgabenRepo) UpdateAufgabeReminderBeforeDue(ctx context.Context, t tx.Tx, taskID string) *app_errors.AppError {
	pgxTx := t.(*tx.PgxTx).Tx
	query := `
	UPDATE aufgaben
	SET reminder_stage = 'Before_Due',
		last_reminder_at = now()
	WHERE id = $1
		AND reminder_stage = 'None';
	`

	if _, err := pgxTx.Exec(ctx, query, taskID); err != nil {
		return app_errors.MapPgxError(err)
	}

	return nil
}

func (r *AufgabenRepo) ListAssignedTasks(ctx context.Context, userID string, filter *aufgaben_dto.AssignedAufgabenFilter) ([]entity.AssignedAufgaben, *app_errors.AppError) {
	query := `
	SELECT a.id, p.name as project_name, a.title, a.description,
	a.status, a.priority, a.due_date
	FROM aufgaben a
	JOIN projects p ON p.id = a.project_id
	WHERE a.assignee_id = $1
		AND a.archived_at IS NULL
	AND (
		$2::aufgaben_status IS NULL OR a.status = $2
	)
	AND (
		$3::aufgaben_priority IS NULL OR a.priority = $3
	)
	AND (
		$4::uuid IS NULL OR a.project_id = $4
	)
	AND (
		$5::uuid IS NULL OR a.id < $5)
	ORDER BY a.due_date ASC, a.id ASC
	LIMIT $6 + 1;
	`

	var aufgaben []entity.AssignedAufgaben
	rows, err := r.db.Query(ctx, query, userID, filter.Status, filter.Priority, filter.ProjectID, filter.Cursor, filter.Limit)
	if err != nil {
		return nil, app_errors.MapPgxError(err)
	}

	defer rows.Close()

	for rows.Next() {
		var aufgabe entity.AssignedAufgaben
		if err := rows.Scan(&aufgabe.ID, &aufgabe.ProjectName, &aufgabe.Title, &aufgabe.Description, &aufgabe.Status, &aufgabe.Priority, &aufgabe.DueDate); err != nil {
			return nil, app_errors.MapPgxError(err)
		}

		aufgaben = append(aufgaben, aufgabe)
	}

	if err := rows.Err(); err != nil {
		return nil, app_errors.MapPgxError(err)
	}

	return aufgaben, nil
}

func (r *AufgabenRepo) ArchiveTask(ctx context.Context, t tx.Tx, taskID string) *app_errors.AppError {
	pgxTx := t.(*tx.PgxTx).Tx
	query := `
	UPDATE aufgaben
	SET status = 'Archived',
		archived_at = now(),
		updated_at = now()
	WHERE id = $1;
	`

	if _, err := pgxTx.Exec(ctx, query, taskID); err != nil {
		return app_errors.MapPgxError(err)
	}
	return nil
}

func (r *AufgabenRepo) UpdateDueDate(ctx context.Context, t tx.Tx, taskID string, dueDate time.Time) (*time.Time, *app_errors.AppError) {
	pgxTx := t.(*tx.PgxTx).Tx
	query := `
	UPDATE aufgaben
	SET due_date = $2,
		reminder_stage = 'None',
		last_reminder_at = NULL,
		updated_at = now()
	WHERE id = $1
		AND due_date IS DISTINCT FROM $2
	RETURNING due_date;
	`

	var updatedDueDate time.Time
	if err := pgxTx.QueryRow(ctx, query, taskID, dueDate).Scan(&updatedDueDate); err != nil {
		return nil, app_errors.MapPgxError(err)
	}

	return &updatedDueDate, nil
}

func (r *AufgabenRepo) ListEventsForTask(ctx context.Context, taskID string, filters *aufgaben_dto.AufgabenEventFilter) ([]entity.AssignmentEventEntity, *app_errors.AppError) {
	query := `
	SELECT id, aufgaben_id, actor_id, target_assignee_id, action, note, reason_code, reason_text, created_at
	FROM aufgaben_assignment_events
	WHERE aufgaben_id = $1
		AND (
		$2::uuid IS NULL OR id < $2
		)
		ORDER BY created_at DESC, id DESC
		LIMIT $3 + 1;
	`

	var events []entity.AssignmentEventEntity
	rows, err := r.db.Query(ctx, query, taskID, filters.Cursor, filters.Limit)

	if err != nil {
		return nil, app_errors.MapPgxError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var event entity.AssignmentEventEntity
		if err := rows.Scan(&event.ID, &event.AufgabenID, &event.ActorID, &event.TargetAssigneeID, &event.Action, &event.Note, &event.ReasonCode, &event.ReasonText, &event.CreatedAt); err != nil {
			return nil, app_errors.MapPgxError(err)
		}
	}

	return events, nil
}

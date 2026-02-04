package aufgaben_repo

import (
	"context"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/abstraction/tx"
	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
)

type AufgabenRepoContract interface {
	CheckProjectMember(ctx context.Context, projectID, userID string) (bool, *app_errors.AppError)
	GetUserRole(ctx context.Context, projectID, userID string) (*entity.UserRole, *app_errors.AppError)
	GetTaskByID(ctx context.Context, taskID string) (*entity.AufgabenEntity, *app_errors.AppError)
	InsertNewAufgaben(ctx context.Context, task *entity.AufgabenEntity) *app_errors.AppError
	CountTasks(ctx context.Context, projectID string) (int64, *app_errors.AppError)
	ListTasks(ctx context.Context, projectID string, filter *aufgaben_dto.AufgabenListFilter) ([]entity.AufgabenEntity, *app_errors.AppError)
	AssignTask(ctx context.Context, t tx.Tx, projectID, taskID, userID string, dueDate *time.Time) (*entity.AssignTaskEntity, *app_errors.AppError)
	ForwardProgress(ctx context.Context, t tx.Tx, taskID string) (*entity.CompleteTaskEntity, *app_errors.AppError)
	InsertAssignmentEvent(ctx context.Context, t tx.Tx, event *entity.AddAssignment) *app_errors.AppError
	UnassignTask(ctx context.Context, t tx.Tx, rollbackModel *entity.UnassignTaskEntity) (entity.AufgabenStatus, *app_errors.AppError)
	ShouldRemind(ctx context.Context, taskID string) (*entity.ReminderAufgaben, *app_errors.AppError)
	UpdateAufgabeReminderBeforeDue(ctx context.Context, t tx.Tx, taskID string) *app_errors.AppError
	ListShouldRemindOverdue(ctx context.Context) ([]entity.ReminderAufgaben, *app_errors.AppError)
	BatchUpdateAufgabenReminderOverdue(ctx context.Context, t tx.Tx, taskIDs []string) *app_errors.AppError
	ListAssignedTasks(ctx context.Context, userID string, filter *aufgaben_dto.AssignedAufgabenFilter) ([]entity.AssignedAufgaben, *app_errors.AppError)
	ArchiveTask(ctx context.Context, t tx.Tx, taskID string) *app_errors.AppError
	UpdateDueDate(ctx context.Context, t tx.Tx, taskID string, dueDate time.Time) (*time.Time, *app_errors.AppError)
	ListEventsForTask(ctx context.Context, taskID string, filters *aufgaben_dto.AufgabenEventFilter) ([]entity.AssignmentEventEntity, *app_errors.AppError)
}

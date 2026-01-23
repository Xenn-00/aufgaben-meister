package aufgaben_repo

import (
	"context"
	"time"

	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/jackc/pgx/v5"
)

type AufgabenRepoContract interface {
	CheckProjectMember(ctx context.Context, projectID, userID string) (bool, *app_errors.AppError)
	GetUserRole(ctx context.Context, projectID, userID string) (*entity.UserRole, *app_errors.AppError)
	GetTaskByID(ctx context.Context, taskID string) (*entity.AufgabenEntity, *app_errors.AppError)
	InsertNewAufgaben(ctx context.Context, task *entity.AufgabenEntity) *app_errors.AppError
	CountTasks(ctx context.Context, projectID string) (int64, *app_errors.AppError)
	ListTasks(ctx context.Context, projectID string, filter *aufgaben_dto.AufgabenListFilter) ([]entity.AufgabenEntity, *app_errors.AppError)
	AssignTask(ctx context.Context, tx pgx.Tx, projectID, taskID, userID string, dueDate *time.Time) (*entity.AssignTaskEntity, *app_errors.AppError)
	ForwardProgress(ctx context.Context, tx pgx.Tx, taskID string) (*entity.CompleteTaskEntity, *app_errors.AppError)
	InsertAssignmentEvent(ctx context.Context, tx pgx.Tx, event *entity.AddAssignment) *app_errors.AppError
	UnassignTask(ctx context.Context, tx pgx.Tx, rollbackModel *entity.UnassignTaskEntity) (entity.AufgabenStatus, *app_errors.AppError)
}

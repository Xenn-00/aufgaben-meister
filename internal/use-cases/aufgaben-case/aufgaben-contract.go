package aufgaben_case

import (
	"context"

	"github.com/Xenn-00/aufgaben-meister/internal/dtos"
	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
)

type AufgabenServiceContract interface {
	CreateNewAufgaben(ctx context.Context, userID, projectID string, req *aufgaben_dto.CreateNewAufgabenRequest) (*aufgaben_dto.CreateNewAufgabenResponse, *app_errors.AppError)
	ListTasksProject(ctx context.Context, userID, projectID string, filter aufgaben_dto.AufgabenListFilter) ([]*aufgaben_dto.AufgabenItem, *dtos.PaginationMeta, *app_errors.AppError)
	AssignTask(ctx context.Context, userID, projectID, taskID string, req *aufgaben_dto.AufgabenAssignRequest) (*aufgaben_dto.AufgabenAssignResponse, *app_errors.AppError)
	GetAufgabeDetails(ctx context.Context, userID, projectID, taskID string) (*aufgaben_dto.AufgabenItem, *app_errors.AppError)
	ForwardProgressTask(ctx context.Context, userID, projectID, taskID string) (*aufgaben_dto.AufgabenForwardProgressResponse, *app_errors.AppError)
	UnassignTask(ctx context.Context, userID, projectID, taskID string, req *aufgaben_dto.UnassignAufgabenRequest) (*aufgaben_dto.UnassignAufgabenResponse, *app_errors.AppError)
	ForceUnassignTask(ctx context.Context, userID, projectID, taskID string, req *aufgaben_dto.ForceUnassignAufgabenRequest) (*aufgaben_dto.UnassignAufgabenResponse, *app_errors.AppError)
	ReassignTask(ctx context.Context, userID, projectID, taskID string, req *aufgaben_dto.ReassignAufgabenRequest) (*aufgaben_dto.ReassignAufgabenResponse, *app_errors.AppError)
	ListAssignedTasks(ctx context.Context, userID string, filter *aufgaben_dto.AssignedAufgabenFilter) ([]*aufgaben_dto.AssignedAufgabenListItem, *dtos.CursorPaginationMeta, *app_errors.AppError)
	ArchiveTask(ctx context.Context, userID, projectID, taskID string) *app_errors.AppError
	UpdateDueDate(ctx context.Context, userID, projectID, taskID string, req *aufgaben_dto.UpdateDueDateRequest) (*aufgaben_dto.UpdateDueDateResponse, *app_errors.AppError)
	FetchEventsForTask(ctx context.Context, userID, projectID, taskID string, filters *aufgaben_dto.AufgabenEventFilter) ([]*aufgaben_dto.AufgabenEventItem, *dtos.CursorPaginationMeta, *app_errors.AppError)
	ForceAufgabeHandover(ctx context.Context, userID, projectID, taskID string, req *aufgaben_dto.ForceAufgabeHandoverRequest) (*aufgaben_dto.ReassignAufgabenResponse, *app_errors.AppError)
}

package aufgaben_case

import (
	"context"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/abstraction/tx"
	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/stretchr/testify/mock"
)

type MockAufgabenRepo struct {
	mock.Mock
}

// Mocking repository that being used in method
func (m *MockAufgabenRepo) CheckProjectMember(ctx context.Context, projectID, userID string) (bool, *app_errors.AppError) {
	args := m.Called(ctx, projectID, userID)
	return args.Bool(0), args.Get(1).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) InsertNewAufgaben(ctx context.Context, task *entity.AufgabenEntity) *app_errors.AppError {
	args := m.Called(ctx, task)
	return args.Get(0).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) GetTaskByID(ctx context.Context, taskID string) (*entity.AufgabenEntity, *app_errors.AppError) {
	args := m.Called(ctx, taskID)
	return args.Get(0).(*entity.AufgabenEntity), args.Get(1).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) CountTasks(ctx context.Context, projectID string) (int64, *app_errors.AppError) {
	args := m.Called(ctx, projectID)
	return int64(args.Int(0)), args.Get(1).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) ListTasks(ctx context.Context, projectID string, filter *aufgaben_dto.AufgabenListFilter) ([]entity.AufgabenEntity, *app_errors.AppError) {
	args := m.Called(ctx, projectID, filter)
	return args.Get(0).([]entity.AufgabenEntity), args.Get(1).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) AssignTask(ctx context.Context, t tx.Tx, projectID, taskID, userID string, dueDate *time.Time) (*entity.AssignTaskEntity, *app_errors.AppError) {
	args := m.Called(ctx, t, projectID, taskID, userID, dueDate)
	return args.Get(0).(*entity.AssignTaskEntity), args.Get(1).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) ForwardProgress(ctx context.Context, t tx.Tx, taskID string) (*entity.CompleteTaskEntity, *app_errors.AppError) {
	args := m.Called(ctx, t, taskID)
	return args.Get(0).(*entity.CompleteTaskEntity), args.Get(1).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) UnassignTask(ctx context.Context, t tx.Tx, rollbackModel *entity.UnassignTaskEntity) (entity.AufgabenStatus, *app_errors.AppError) {
	args := m.Called(ctx, t, rollbackModel)
	return args.Get(0).(entity.AufgabenStatus), args.Get(1).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) ShouldRemind(ctx context.Context, taskID string) (*entity.ReminderAufgaben, *app_errors.AppError) {
	args := m.Called(ctx, taskID)
	return args.Get(0).(*entity.ReminderAufgaben), args.Get(1).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) UpdateAufgabeReminderBeforeDue(ctx context.Context, t tx.Tx, taskID string) *app_errors.AppError {
	args := m.Called(ctx, t, taskID)
	return args.Get(0).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) ListShouldRemindOverdue(ctx context.Context) ([]entity.ReminderAufgaben, *app_errors.AppError) {
	args := m.Called(ctx)
	return args.Get(0).([]entity.ReminderAufgaben), args.Get(1).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) ArchiveTask(ctx context.Context, t tx.Tx, taskID string) *app_errors.AppError {
	args := m.Called(ctx, t, taskID)
	return args.Get(0).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) BatchUpdateAufgabenReminderOverdue(ctx context.Context, t tx.Tx, taskIDs []string) *app_errors.AppError {
	args := m.Called(ctx, t, taskIDs)
	return args.Get(0).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) InsertAssignmentEvent(ctx context.Context, t tx.Tx, event *entity.AddAssignment) *app_errors.AppError {
	args := m.Called(ctx, t, event)
	return args.Get(0).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) ListAssignedTasks(ctx context.Context, userID string, filter *aufgaben_dto.AssignedAufgabenFilter) ([]entity.AssignedAufgaben, *app_errors.AppError) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).([]entity.AssignedAufgaben), args.Get(1).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) UpdateDueDate(ctx context.Context, t tx.Tx, taskID string, dueDate time.Time) (*time.Time, *app_errors.AppError) {
	args := m.Called(ctx, t, taskID, dueDate)
	return args.Get(0).(*time.Time), args.Get(1).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) ListEventsForTask(ctx context.Context, taskID string, filters *aufgaben_dto.AufgabenEventFilter) ([]entity.AssignmentEventEntity, *app_errors.AppError) {
	args := m.Called(ctx, taskID, filters)
	return args.Get(0).([]entity.AssignmentEventEntity), args.Get(1).(*app_errors.AppError)
}

func (m *MockAufgabenRepo) GetUserRole(ctx context.Context, projectID, userID string) (*entity.UserRole, *app_errors.AppError) {
	args := m.Called(ctx, projectID, userID)
	return args.Get(0).(*entity.UserRole), args.Get(1).(*app_errors.AppError)
}

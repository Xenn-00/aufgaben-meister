package aufgaben_case

import (
	"context"
	"testing"
	"time"

	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/stretchr/testify/assert"
)

// Test Cache hit
func TestGetAufgabeDetails_CacheHit(t *testing.T) {
	ctx := context.Background()
	repo := new(MockAufgabenRepo)

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	cache := &MockCache{
		GetFn: func(ctx context.Context, key string) (*any, *app_errors.AppError) {
			var v any = &aufgaben_dto.AufgabenItem{
				AufgabenID: "task-1",
				Title:      "Cache Task",
			}
			return &v, nil
		},
	}

	service := &AufgabenService{repo: repo, cache: cache}

	resp, err := service.GetAufgabeDetails(ctx,
		userID, projectID, taskID)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Cache Task", resp.Title)
	assert.Equal(t, 1, cache.GetCalled)
}

// Test Forbidden and Cache hit
func TestGetAufgabeDetails_ForbiddenCacheHit(t *testing.T) {
	ctx := context.Background()
	repo := new(MockAufgabenRepo)

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(false, (*app_errors.AppError)(nil))

	cache := &MockCache{
		GetFn: func(ctx context.Context, key string) (*any, *app_errors.AppError) {
			var v any = &aufgaben_dto.AufgabenItem{
				AufgabenID: "task-1",
				Title:      "Cache Task",
			}
			return &v, nil
		},
	}

	service := &AufgabenService{repo: repo, cache: cache}

	resp, err := service.GetAufgabeDetails(ctx,
		userID, projectID, taskID)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)

	repo.AssertExpectations(t)
}

// Test Cache miss
func TestGetAufgabeDetails_CacheMiss(t *testing.T) {
	ctx := context.Background()
	repo := new(MockAufgabenRepo)

	userID := "user-1"
	projectID := "project-1"
	taskID := "task-1"

	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	cache := &MockCache{
		GetFn: func(ctx context.Context, key string) (*any, *app_errors.AppError) {
			return nil, nil
		},
		SetFn: func(ctx context.Context, key string, val *any, ttl time.Duration) *app_errors.AppError {
			return nil
		},
	}

	service := &AufgabenService{repo: repo, cache: cache}

	taskDescription := "Important"
	dueDate := time.Now()
	assigneeID := "user-2"
	r := &entity.AufgabenEntity{
		ID:          "task-1",
		Title:       "Test Task",
		Description: &taskDescription,
		Status:      entity.AufgabenInProgress,
		Priority:    entity.PriorityMedium,
		AssigneeID:  &assigneeID,
		DueDate:     &dueDate,
	}

	repo.On("GetTaskByID", ctx, taskID).Return((*entity.AufgabenEntity)(r), (*app_errors.AppError)(nil))

	resp, err := service.GetAufgabeDetails(ctx, userID, projectID, taskID)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Test Task", resp.Title)

	assert.Equal(t, 1, cache.GetCalled)
	assert.Equal(t, 1, cache.SetCalled)

	repo.AssertExpectations(t)
}

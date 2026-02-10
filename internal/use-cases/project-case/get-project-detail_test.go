package project_case

import (
	"context"
	"testing"
	"time"

	project_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/project-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	use_cases "github.com/Xenn-00/aufgaben-meister/internal/use-cases"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Test cache hit
func TestGetProjectDetail_CacheHit_Meister(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	projectID := "project-1"
	userID := "meister-1"
	meisterRole := string(entity.MEISTER)

	// Setup expectations
	repo.On("GetUserRoleInProject", ctx, userID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	// Cache hit - return cached data
	cache := &use_cases.MockCache{
		GetFn: func(ctx context.Context, key string) (*any, *app_errors.AppError) {
			masterID := "master-123"
			var v any = &project_dto.GetProjectDetailResponse{
				ID:          projectID,
				Name:        "Cached Project",
				TypeProject: string(entity.PERSONAL),
				Visibility:  string(entity.PRIVATE),
				Role:        &meisterRole,
				MasterID:    &masterID,
				Members: []entity.ProjectMember{
					{UserID: "user-1", Username: "cached_user", Role: entity.MITARBEITER},
				},
			}
			return &v, nil
		},
	}

	service := &ProjectService{repo: repo, cache: cache}

	// Execute
	resp, err := service.GetProjectDetail(ctx, projectID, userID)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Cached Project", resp.Name)
	assert.Equal(t, 1, cache.GetCalled)
	assert.Equal(t, 0, cache.SetCalled) // Cache hit, no set

	repo.AssertExpectations(t)
}

// Test 2: Cache Miss - MEISTER gets project detail and sets cache
func TestGetProjectDetail_CacheMiss_Meister(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	projectID := "project-1"
	userID := "meister-1"
	meisterRole := string(entity.MEISTER)

	// Setup expectations
	repo.On("GetUserRoleInProject", ctx, userID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	// Cache miss - return nil
	cache := &use_cases.MockCache{
		GetFn: func(ctx context.Context, key string) (*any, *app_errors.AppError) {
			return nil, nil
		},
		SetFn: func(ctx context.Context, key string, val *any, ttl time.Duration) *app_errors.AppError {
			return nil
		},
	}

	// GetProjectByID
	masterID := "master-123"
	project := &entity.ProjectEntity{
		ID:         projectID,
		Name:       "Test Project",
		Type:       entity.PERSONAL,
		Visibility: entity.PRIVATE,
		MasterID:   masterID,
	}

	repo.On("GetProjectByID", ctx, projectID).Return(project, (*app_errors.AppError)(nil))

	// GetProjectMember (only for MEISTER)
	members := []entity.ProjectMember{
		{
			UserID:   "user-1",
			Username: "john_doe",
			Role:     entity.MITARBEITER,
		},
	}

	repo.On("GetProjectMember", ctx, projectID).Return(members, (*app_errors.AppError)(nil))

	service := &ProjectService{repo: repo, cache: cache}

	// Execute
	resp, err := service.GetProjectDetail(ctx, projectID, userID)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, projectID, resp.ID)
	assert.Equal(t, "Test Project", resp.Name)
	assert.Equal(t, &meisterRole, resp.Role)
	assert.Equal(t, &masterID, resp.MasterID)
	assert.NotNil(t, resp.Members)
	assert.Len(t, resp.Members, 1)

	assert.Equal(t, 1, cache.GetCalled) // Cache miss, check cache
	assert.Equal(t, 1, cache.SetCalled) // Cache miss, set cache

	repo.AssertExpectations(t)
}

// Test 3: Cache Miss - MITARBEITER gets project detail without members
func TestGetProjectDetail_CacheMiss_Mitarbeiter(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	projectID := "project-1"
	userID := "mitarbeiter-1"
	mitarbeiterRole := string(entity.MITARBEITER)

	// Setup expectations
	repo.On("GetUserRoleInProject", ctx, userID, projectID).Return(mitarbeiterRole, (*app_errors.AppError)(nil))

	// Cache miss
	cache := &use_cases.MockCache{
		GetFn: func(ctx context.Context, key string) (*any, *app_errors.AppError) {
			return nil, nil
		},
		SetFn: func(ctx context.Context, key string, val *any, ttl time.Duration) *app_errors.AppError {
			return nil
		},
	}

	// GetProjectByID
	masterID := "master-123"
	project := &entity.ProjectEntity{
		ID:         projectID,
		Name:       "Test Project",
		Type:       entity.CORPORATE,
		Visibility: entity.PRIVATE,
		MasterID:   masterID,
	}

	repo.On("GetProjectByID", ctx, projectID).Return(project, (*app_errors.AppError)(nil))

	// GetProjectMember should NOT be called for MITARBEITER

	service := &ProjectService{repo: repo, cache: cache}

	// Execute
	resp, err := service.GetProjectDetail(ctx, projectID, userID)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, projectID, resp.ID)
	assert.Equal(t, &mitarbeiterRole, resp.Role)
	assert.Nil(t, resp.MasterID) // MITARBEITER does NOT get MasterID
	assert.Nil(t, resp.Members)  // MITARBEITER does NOT get Members

	assert.Equal(t, 1, cache.GetCalled)
	assert.Equal(t, 1, cache.SetCalled)

	repo.AssertExpectations(t)
}

// Test 4: User has no role in project (not a member)
func TestGetProjectDetail_UserNotProjectMember(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	cache := &use_cases.MockCache{
		GetFn: func(ctx context.Context, key string) (*any, *app_errors.AppError) {
			return nil, nil
		},
	}

	service := &ProjectService{repo: repo, cache: cache}

	projectID := "project-1"
	userID := "user-999"

	// Setup expectations
	// GetUserRoleInProject returns error (user not member)
	notMemberError := app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "user_not_member", nil)
	repo.On("GetUserRoleInProject", ctx, userID, projectID).Return("", notMemberError)

	// Execute
	resp, err := service.GetProjectDetail(ctx, projectID, userID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, notMemberError, err)
	assert.Equal(t, 0, cache.GetCalled) // Should not check cache if not a member

	repo.AssertExpectations(t)
}

// Test 5: Project not found
func TestGetProjectDetail_ProjectNotFound(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	projectID := "project-999"
	userID := "user-1"
	meisterRole := string(entity.MEISTER)

	cache := &use_cases.MockCache{
		GetFn: func(ctx context.Context, key string) (*any, *app_errors.AppError) {
			return nil, nil // Cache miss
		},
	}

	service := &ProjectService{repo: repo, cache: cache}

	// Setup expectations
	repo.On("GetUserRoleInProject", ctx, userID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	// GetProjectByID returns not found error
	notFoundError := app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
	repo.On("GetProjectByID", ctx, projectID).Return((*entity.ProjectEntity)(nil), notFoundError)

	// Execute
	resp, err := service.GetProjectDetail(ctx, projectID, userID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, notFoundError, err)

	repo.AssertExpectations(t)
}

// Test 6: GetProjectMember fails for MEISTER
func TestGetProjectDetail_GetProjectMemberFails(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	projectID := "project-1"
	userID := "meister-1"
	meisterRole := string(entity.MEISTER)

	cache := &use_cases.MockCache{
		GetFn: func(ctx context.Context, key string) (*any, *app_errors.AppError) {
			return nil, nil // Cache miss
		},
	}

	service := &ProjectService{repo: repo, cache: cache}

	// Setup expectations
	repo.On("GetUserRoleInProject", ctx, userID, projectID).Return(meisterRole, (*app_errors.AppError)(nil))

	masterID := "master-123"
	project := &entity.ProjectEntity{
		ID:         projectID,
		Name:       "Test Project",
		Type:       entity.PERSONAL,
		Visibility: entity.PRIVATE,
		MasterID:   masterID,
	}

	repo.On("GetProjectByID", ctx, projectID).Return(project, (*app_errors.AppError)(nil))

	// GetProjectMember fails
	memberError := app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "database_error", nil)
	repo.On("GetProjectMember", ctx, projectID).Return(([]entity.ProjectMember)(nil), memberError)

	// Execute
	resp, err := service.GetProjectDetail(ctx, projectID, userID)

	// Assert
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, memberError, err)

	repo.AssertExpectations(t)
}

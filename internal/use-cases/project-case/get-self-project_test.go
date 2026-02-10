package project_case

import (
	"context"
	"testing"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Test happy path
func TestGetSelfProject_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{
		repo: repo,
	}

	userID := "user-1"

	mockProjects := []entity.ProjectSelf{
		{
			ID:         "project-1",
			Name:       "Personal Project",
			Type:       "Personal",
			Visibility: "Private",
			MasterID:   userID,
			Role:       "Meister",
		},
		{
			ID:         "project-2",
			Name:       "Team Project",
			Type:       "Community",
			Visibility: "Public",
			MasterID:   "user-2",
			Role:       "Mitarbeiter",
		},
		{
			ID:         "project-3",
			Name:       "Coorporate Project",
			Type:       "Coorporate",
			Visibility: "Private",
			MasterID:   "user-3",
			Role:       "Mitarbeiter",
		},
	}

	repo.On("GetSelfProject", ctx, userID).Return(mockProjects, (*app_errors.AppError)(nil))

	resp, err := service.GetSelfProject(ctx, userID)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 3)

	assert.Equal(t, "project-1", resp[0].ID)
	assert.Equal(t, "Personal Project", resp[0].Name)
	assert.Equal(t, "Personal", resp[0].TypeProject)
	assert.Equal(t, "Private", resp[0].Visibility)
	assert.Equal(t, "Meister", resp[0].Role)

	assert.Equal(t, "project-2", resp[1].ID)
	assert.Equal(t, "Team Project", resp[1].Name)
	assert.Equal(t, "Mitarbeiter", resp[1].Role)

	assert.Equal(t, "project-3", resp[2].ID)
	assert.Equal(t, "Coorporate Project", resp[2].Name)
	assert.Equal(t, "Mitarbeiter", resp[2].Role)

	repo.AssertExpectations(t)
}

// Test empty projects
func TestGetSelfProject_EmptyProjects(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{
		repo: repo,
	}

	userID := "user-no-project"

	mockProjects := []entity.ProjectSelf{}

	repo.On("GetSelfProject", ctx, userID).Return(mockProjects, (*app_errors.AppError)(nil))

	resp, err := service.GetSelfProject(ctx, userID)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 0)
	assert.Empty(t, resp)

	repo.AssertExpectations(t)
}

// Test single project
func TestGetSelfProject_SingleProject(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{
		repo: repo,
	}

	userID := "user-1"

	mockProjects := []entity.ProjectSelf{
		{
			ID:         "project-solo",
			Name:       "Solo Project",
			Type:       "Personal",
			Visibility: "Private",
			MasterID:   userID,
			Role:       "Meister",
		},
	}

	repo.On("GetSelfProject", ctx, userID).Return(mockProjects, (*app_errors.AppError)(nil))

	resp, err := service.GetSelfProject(ctx, userID)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, "project-solo", resp[0].ID)
	assert.Equal(t, "Solo Project", resp[0].Name)
	assert.Equal(t, "Meister", resp[0].Role)

	repo.AssertExpectations(t)
}

// Test repo error
func TestGetSelfProject_RepositoryError(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{
		repo: repo,
	}

	userID := "user-1"

	expectedErr := app_errors.NewAppError(
		fiber.StatusInternalServerError,
		app_errors.ErrInternal,
		"database_error",
		nil,
	)

	repo.On("GetSelfProject", ctx, userID).Return(
		([]entity.ProjectSelf)(nil),
		expectedErr,
	)

	resp, err := service.GetSelfProject(ctx, userID)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, err.Code)
	assert.Equal(t, app_errors.ErrInternal, err.Type)
	assert.Equal(t, "database_error", err.MessageKey)

	repo.AssertExpectations(t)
}

// Test User Not Found
func TestGetSelfProject_UserNotFound(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	service := &ProjectService{
		repo: repo,
	}

	userID := "non-existent-user"

	expectedErr := app_errors.NewAppError(
		fiber.StatusNotFound,
		app_errors.ErrNotFound,
		"user_not_found",
		nil,
	)

	repo.On("GetSelfProject", ctx, userID).Return(
		([]entity.ProjectSelf)(nil),
		expectedErr,
	)

	resp, err := service.GetSelfProject(ctx, userID)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusNotFound, err.Code)
	assert.Equal(t, app_errors.ErrNotFound, err.Type)
	assert.Equal(t, "user_not_found", err.MessageKey)

	repo.AssertExpectations(t)
}

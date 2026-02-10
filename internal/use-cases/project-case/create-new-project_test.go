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
	"github.com/stretchr/testify/mock"
)

// Test happy path
func TestCreateNewProject_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "user-1"
	projectID := "project-1"

	req := &project_dto.CreateNewProjectRequest{
		Name:        "Test Project",
		TypeProject: "Personal",
		Visibility:  "Private",
	}

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	project := &entity.ProjectEntity{
		ID:         "project-1",
		Name:       req.Name,
		Type:       entity.ProjectType("Personal"),
		Visibility: entity.ProjectVisibility("Private"),
		MasterID:   userID,
		CreatedAt:  time.Now(),
	}

	repo.On("InsertNewProject", ctx, tx, mock.Anything).Return(project, (*app_errors.AppError)(nil))

	// Project creator automatically being project meister
	repo.On("InsertNewProjectMember", ctx, tx, projectID, userID, entity.MEISTER).Return((*app_errors.AppError)(nil))

	tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	resp, err := service.CreateNewProject(ctx, req, userID)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	// assert.Equal(t, "project-1", resp.ID)
	assert.Equal(t, "Test Project", resp.Name)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

// Test transaction failed
func TestCreateNewProject_TransactionBeginFailed(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "user-1"

	req := &project_dto.CreateNewProjectRequest{
		Name:        "Test Project",
		TypeProject: "Personal",
		Visibility:  "Private",
	}

	expectedErr := app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", nil)
	txManager.On("Begin", ctx).Return((*use_cases.MockTx)(nil), expectedErr)

	resp, err := service.CreateNewProject(ctx, req, userID)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, err.Code)
	assert.Equal(t, app_errors.ErrInternal, err.Type)

	txManager.AssertExpectations(t)
}

func TestCreateNewProject_InsertProjectMemberFailed(t *testing.T) {
	ctx := context.Background()

	repo := new(MockProjectRepo)
	txManager := new(use_cases.MockTxManager)
	tx := new(use_cases.MockTx)
	service := &ProjectService{
		repo:      repo,
		txManager: txManager,
	}

	userID := "user-1"

	req := &project_dto.CreateNewProjectRequest{
		Name:        "Test Project",
		TypeProject: "Personal",
		Visibility:  "Private",
	}

	txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

	repo.On("InsertNewProject", ctx, tx, mock.Anything).Return(
		&entity.ProjectEntity{
			ID:         "project-1",
			Name:       req.Name,
			Type:       entity.ProjectType("Personal"),
			Visibility: entity.ProjectVisibility("Private"),
			MasterID:   userID,
			CreatedAt:  time.Now(),
		},
		(*app_errors.AppError)(nil),
	)

	expectedErr := app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "failed_insert_member", nil)
	repo.On("InsertNewProjectMember", ctx, tx, "project-1", userID, entity.MEISTER).Return(expectedErr)

	tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

	resp, err := service.CreateNewProject(ctx, req, userID)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, err.Code)
	assert.Equal(t, "failed_insert_member", err.MessageKey)

	repo.AssertExpectations(t)
	tx.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

func TestCreateNewProject_DifferentTypesAndVisibility(t *testing.T) {
	testCases := []struct {
		name        string
		projectType string
		visibility  string
	}{
		{"Team Public Project", "Team", "Public"},
		{"Personal Private Project", "Personal", "Private"},
		{"Coorporate Internal Project", "Coorporate", "Internal"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			repo := new(MockProjectRepo)
			txManager := new(use_cases.MockTxManager)
			tx := new(use_cases.MockTx)
			service := &ProjectService{
				repo:      repo,
				txManager: txManager,
			}

			userID := "user-1"

			req := &project_dto.CreateNewProjectRequest{
				Name:        tc.name,
				TypeProject: tc.projectType,
				Visibility:  tc.visibility,
			}

			txManager.On("Begin", ctx).Return(tx, (*app_errors.AppError)(nil))

			repo.On("InsertNewProject", ctx, tx, mock.Anything).Return(
				&entity.ProjectEntity{
					ID:         "project-test",
					Name:       req.Name,
					Type:       entity.ProjectType(tc.projectType),
					Visibility: entity.ProjectVisibility(tc.visibility),
					MasterID:   userID,
					CreatedAt:  time.Now(),
				},
				(*app_errors.AppError)(nil),
			)

			repo.On("InsertNewProjectMember", ctx, tx, "project-test", userID, entity.MEISTER).Return((*app_errors.AppError)(nil))

			tx.On("Commit", ctx).Return((*app_errors.AppError)(nil))
			tx.On("Rollback", ctx).Return((*app_errors.AppError)(nil))

			resp, err := service.CreateNewProject(ctx, req, userID)

			assert.Nil(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tc.name, resp.Name)
			assert.Equal(t, tc.projectType, resp.TypeProject)
			assert.Equal(t, tc.visibility, resp.Visibility)

			repo.AssertExpectations(t)
			tx.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

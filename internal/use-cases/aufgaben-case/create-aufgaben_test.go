package aufgaben_case

import (
	"context"
	"testing"

	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test Happy path
func TestCreateNewAufgaben_Success(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)
	service := &AufgabenService{
		repo: repo,
	}

	userID := "user-1"
	projectID := "project-1"

	taskDescription := "Do Something"
	req := &aufgaben_dto.CreateNewAufgabenRequest{
		Title:       "Test Task",
		Description: &taskDescription,
	}

	// expectation
	repo.On("CheckProjectMember", ctx, projectID, userID).Return(true, (*app_errors.AppError)(nil))

	repo.On("InsertNewAufgaben", ctx, mock.AnythingOfType("*entity.AufgabenEntity")).Return((*app_errors.AppError)(nil))

	resp, err := service.CreateNewAufgaben(ctx, userID, projectID, req)

	assert.Nil(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, req.Title, resp.Title)
	assert.Equal(t, projectID, resp.ProjectID)

	repo.AssertExpectations(t)
}

// Test performer not project member
func TestCreateNewAufgaben_PerformerNotProjectMember(t *testing.T) {
	ctx := context.Background()

	repo := new(MockAufgabenRepo)

	service := &AufgabenService{
		repo: repo,
	}

	repo.On("CheckProjectMember", ctx, "project-1", "user-1").Return(false, (*app_errors.AppError)(nil))

	resp, err := service.CreateNewAufgaben(ctx, "user-1", "project-1", &aufgaben_dto.CreateNewAufgabenRequest{
		Title: "Task",
	})

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, app_errors.ErrForbidden, err.Type)

	repo.AssertExpectations(t)
}

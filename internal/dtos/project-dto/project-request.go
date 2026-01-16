package project_dto

import "github.com/go-playground/validator/v10"

type CreateNewProjectRequest struct {
	Name        string `json:"name_project" validate:"required,min=3,max=255"`
	TypeProject string `json:"type_project" validate:"required,typeProject"`
	Visibility  string `json:"project_visibility" validate:"required,visibility"`
}

type ParamGetProjectByID struct {
	ID string `params:"project_id" validate:"required,uuid"`
}

type InviteProjectMemberRequest struct {
	UserIDs []string `json:"user_ids" validate:"required,min=1,dive,uuid"`
}

type InvitationQueryRequest struct {
	InvitationID string `query:"invitation_id" validate:"required,uuid"`
	Token        string `query:"token" validate:"required,max=21"`
}

type ProjectType string

const (
	ProjectTypePersonal  ProjectType = "Personal"
	ProjectTypeCommunity ProjectType = "Community"
	ProjectTypeCorporate ProjectType = "Corporate"
)

func IsValidTypeProject(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	switch ProjectType(v) {
	case ProjectTypePersonal, ProjectTypeCommunity, ProjectTypeCorporate:
		return true
	default:
		return false
	}
}

type ProjectVisibility string

const (
	VisibilityPublic  ProjectVisibility = "Public"
	VisibilityPrivate ProjectVisibility = "Private"
)

func IsValidVisibility(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	switch ProjectVisibility(v) {
	case VisibilityPublic, VisibilityPrivate:
		return true
	default:
		return false
	}
}

package project_dto

import (
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
)

type CreateNewProjectResponse struct {
	ID          string `json:"project_id"`
	Name        string `json:"project_name"`
	MasterID    string `json:"project_master"`
	TypeProject string `json:"type_project"`
	Visibility  string `json:"visibility"`
}

type SelfProjectResponse struct {
	ID          string `json:"project_id"`
	Name        string `json:"project_name"`
	MasterID    string `json:"project_master"`
	TypeProject string `json:"type_project"`
	Visibility  string `json:"visibility"`
	Role        string `json:"role"`
}

type GetProjectDetailResponse struct {
	ID          string                 `json:"project_id"`
	Name        string                 `json:"project_name"`
	MasterID    *string                `json:"project_master,omitempty"`
	TypeProject string                 `json:"type_project"`
	Visibility  string                 `json:"visibility"`
	Role        *string                `json:"role,omitempty"`
	Members     []entity.ProjectMember `json:"members"`
}

type InviteProjectMemberResponse struct {
	Invited      []string      `json:"invited"`
	SkippedUsers []SkippedUser `json:"skipped_users"`
}

type InvitationMemberAccepted struct {
	ID         string    `json:"project_id"`
	Name       string    `json:"project_name"`
	Role       string    `json:"role"`
	AcceptedAt time.Time `json:"accepted_at"`
}

type SkippedUser struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"` // already_member | already_invited | user_not_found
}

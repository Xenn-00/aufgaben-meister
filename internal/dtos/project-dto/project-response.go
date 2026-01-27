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

type SelfProjectInvitationResponse struct {
	ID          string    `json:"invitation_id"`
	ProjectID   string    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expires_at"`
	InvitedBy   string    `json:"invited_by"` // better filled with username or name who's being inviting
}

type RejectProjectInvitationResponse struct {
	ID        string    `json:"invitation_id"`
	ProjectID string    `json:"project_id"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
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

type RevokeProjectMemberResponse struct {
	Revoked      []string      `json:"Revoked"`
	RevokedUsers []RevokedUser `json:"revoked_users"`
}

// type ListInvitationsResponse struct {
// 	Data       []*InvitationsInProjectResponse `json:"data"`
// 	NextCursor *time.Time                      `json:"next_cursor"`
// 	HasMore    bool                            `json:"has_more"`
// }

type InvitationsInProjectResponse struct {
	InvitationID string    `json:"invitation_id"`
	ProjectName  string    `json:"project_name"`
	UserID       string    `json:"user_id"`
	Status       string    `json:"status"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type RevokedUser struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
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

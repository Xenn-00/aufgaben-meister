package entity

import (
	"time"
)

type ProjectEntity struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       ProjectType       `json:"type"`
	Visibility ProjectVisibility `json:"visibility"`
	MasterID   string            `json:"master_id"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type ProjectMember struct {
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Role      UserRole  `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

type ProjectSelf struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Visibility string `json:"visibility"`
	Role       string `json:"role"`
	MasterID   string `json:"master_id"`
}

type ProjectInvitationEntity struct {
	ID            string               `json:"id"`
	ProjectID     string               `json:"project_id"`
	ProjectName   string               `json:"project_name"`
	InvitedUserID string               `json:"invited_user_id"`
	InvitedBy     string               `json:"invited_by"`
	Role          UserRole             `json:"role"`
	Status        InvitationStatusEnum `json:"status"`
	TokenHash     string               `json:"token_hash"`
	ExpiresAt     time.Time            `json:"expires_at"`
	AcceptedAt    *time.Time           `json:"accepted_at,omitempty"`
	RevokedAt     *time.Time           `json:"revoked_at,omitempty"`
	RejectedAt    *time.Time           `json:"rejected_at,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
}

type InvitationInfo struct {
	ID               string `json:"id"`
	InvitationStatus string `json:"status"`
	ProjectID        string `json:"project_id"`
	ProjectName      string `json:"project_name"`
	Username         string `json:"username"`
	UserEmail        string `json:"email"`
}

type ProjectType string
type ProjectVisibility string
type InvitationStatusEnum string

const (
	PERSONAL  ProjectType = "Personal"
	COMMUNITY ProjectType = "Community"
	CORPORATE ProjectType = "Corporate"

	PUBLIC  ProjectVisibility = "Public"
	PRIVATE ProjectVisibility = "Private"

	PENDING  InvitationStatusEnum = "Pending"
	ACCEPTED InvitationStatusEnum = "Accepted"
	REVOKED  InvitationStatusEnum = "Revoked"
	EXPIRED  InvitationStatusEnum = "Expired"
	REJECTED InvitationStatusEnum = "Rejected"
)

func (s ProjectType) IsValid() bool {
	switch s {
	case PERSONAL, COMMUNITY, CORPORATE:
		return true
	}
	return false
}

func (s ProjectVisibility) IsValid() bool {
	switch s {
	case PUBLIC, PRIVATE:
		return true
	}

	return false
}

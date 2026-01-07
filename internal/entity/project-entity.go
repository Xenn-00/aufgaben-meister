package entity

import (
	"time"
)

type ProjectEntity struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       ProjectType       `json:"project_type"`
	Visibility ProjectVisibility `json:"project_visibility"`
	MasterID   string            `json:"master_id"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type ProjectMember struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	Role      UserRole  `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

type ProjectType string
type ProjectVisibility string

const (
	PERSONAL  ProjectType = "Personal"
	COMMUNITY ProjectType = "Community"
	CORPORATE ProjectType = "Corporate"

	PUBLIC  ProjectVisibility = "Public"
	PRIVATE ProjectVisibility = "Private"
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

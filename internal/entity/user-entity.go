package entity

import "time"

// UserEntity repräsentiert die Benutzerdaten in der Datenbank.
type UserEntity struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserCountFilter repräsentiert die Filterkriterien für die Zählung von Benutzern.
type UserCountFilter struct {
	Email    *string
	Username *string
}

type UserWithProject struct {
	ID       string        `json:"id"`
	Email    string        `json:"email"`
	Username string        `json:"username"`
	Name     string        `json:"name"`
	Projects []UserProject `json:"user_projects"`
}
type UserProject struct {
	ProjectID         string            `json:"project_id"`
	ProjectName       string            `json:"project_name"`
	ProjectType       ProjectType       `json:"project_type"`
	ProjectVisibility ProjectVisibility `json:"project_visibility"`
	JoinedAtProject   time.Time         `json:"joined_at_project"`
}

type UserUpdate struct {
	ID        string `json:"ID"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	UpdatedAt string `json:"updated_at"`
}

type UserRole string

const (
	MITARBEITER UserRole = "Mitarbeiter"
	MEISTER     UserRole = "Meister"
)

func (u UserRole) IsValid() bool {
	switch u {
	case MITARBEITER, MEISTER:
		return true
	}

	return false
}

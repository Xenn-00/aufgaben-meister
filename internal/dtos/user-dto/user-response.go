package user_dto

import (
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
)

type UserProfileResponse struct {
	ID        string               `json:"id"`
	Email     string               `json:"email,omitempty"`
	Username  string               `json:"username"`
	Name      string               `json:"name"`
	Project   []entity.UserProject `json:"user_projects,omitempty"`
	CreatedAt time.Time            `json:"created_at,omitzero"`
	UpdatedAt time.Time            `json:"updated_at,omitzero"`
}

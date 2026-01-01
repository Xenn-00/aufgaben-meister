package entity

import "time"

// UserEntity repräsentiert die Benutzerdaten in der Datenbank.
type UserEntity struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserCountFilter repräsentiert die Filterkriterien für die Zählung von Benutzern.
type UserCountFilter struct {
	Email    *string
	Username *string
}

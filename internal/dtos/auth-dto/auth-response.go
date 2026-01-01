package auth_dto

import "time"

// AuthResponse repräsentiert die Daten, die in Authentifizierungsantworten zurückgegeben werden.

// ResisterUserResponse repräsentiert die Daten, die nach der Registrierung eines Benutzers zurückgegeben werden.
type RegisterUserResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

// LoginUserResponse repräsentiert die Daten, die nach der Anmeldung eines Benutzers zurückgegeben werden.
type LoginUserResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

type ListAllUserDevicesResponse struct {
	Key       string    `json:"key"`
	Device    string    `json:"device"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	LoginAt   time.Time `json:"login_at"`
}

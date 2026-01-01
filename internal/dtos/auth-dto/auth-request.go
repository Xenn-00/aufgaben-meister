package auth_dto

// AuthRequest repräsentiert die Daten, die für Authentifizierungsanfragen benötigt werden.

// RegisterUserRequest repräsentiert die Daten, die für die Registrierung eines Benutzers benötigt werden.
type RegisterUserRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Name            string `json:"name" validate:"required,min=3"`
	Username        string `json:"username" validate:"required,min=3,max=30"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// LoginUserRequest repräsentiert die Daten, die für die Anmeldung eines Benutzers benötigt werden
type LoginUserRequest struct {
	Identifier string `json:"username_or_email" validate:"required"` // Es könnte sich um eine E-mail oder einen Benutzernamen handeln.
	Password   string `json:"password" validate:"required"`
}

type LoginMetadata struct {
	UserAgent string
	Device    string
	IP        string
}

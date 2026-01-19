package app_errors

// AppError repräsentiert einen Anwendungsfehler mit einem Code, einer Nachricht und optional einem Feld.
type AppError struct {
	Code       int          // HTTP status code
	Type       string       // VALIDATION_ERROR, NOT_FOUND, usw
	MessageKey string       // i18n key
	Details    []FieldError // optional (validation)
	Err        error        // original error (internal only)
}

const (
	ErrValidation   = "VALIDATION_ERROR"
	ErrInvalidBody  = "INVALID_BODY"
	ErrInvalidParam = "INVALID_PARAM"
	ErrInvalidQuery = "INVALID_QUERY"
	ErrUnauthorized = "UNAUTHORIZED"
	ErrForbidden    = "FORBIDDEN"
	ErrNotFound     = "NOT_FOUND"
	ErrConflict     = "CONFLICT"
	ErrInternal     = "INTERNAL_ERROR"
)

type FieldError struct {
	Field      string         `json:"field"`
	Reason     string         `json:"reason"`
	MessageKey string         `json:"message_key"`
	Params     map[string]any `json:"params,omitempty"`
}

func NewAppError(code int, errType string, messageKey string, err error) *AppError {
	return &AppError{
		Code:       code,
		Type:       errType,
		MessageKey: messageKey,
		Err:        err,
	}
}
func NewValidationError(details []FieldError) *AppError {
	return &AppError{
		Code:       400,
		Type:       ErrValidation,
		MessageKey: "invalid_request",
		Details:    details,
	}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.MessageKey
}

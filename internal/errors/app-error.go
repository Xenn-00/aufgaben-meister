package app_errors

import "fmt"

// AppError repräsentiert einen Anwendungsfehler mit einem Code, einer Nachricht und optional einem Feld.
type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e AppError) Error() string {
	return e.Message
}

func New(code int, message, field string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Field:   field,
	}
}
func Newf(code int, format, field string, args ...any) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Field:   field,
	}
}

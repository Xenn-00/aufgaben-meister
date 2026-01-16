package dtos

// WebResponse repräsentiert eine standardisierte Webantwort.
type WebResponse[T any] struct {
	Message   string         `json:"message"`
	Data      T              `json:"data"`
	Details   []T            `json:"details,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
	Errors    *ErrorResponse `json:"errors,omitempty"`
}

// ErrorResponse repräsentiert eine standardisierte Fehlerantwort.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

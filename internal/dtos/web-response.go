package dtos

// WebResponse repräsentiert eine standardisierte Webantwort.
type WebResponse[T any] struct {
	Message   string         `json:"message"`
	Data      T              `json:"data"`
	Details   []any          `json:"details,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
	Errors    *ErrorResponse `json:"errors,omitempty"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ErrorResponse repräsentiert eine standardisierte Fehlerantwort.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

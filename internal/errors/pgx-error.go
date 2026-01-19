package app_errors

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func MapPgxError(err error) *AppError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return NewAppError(409, ErrConflict, "conflict", err)
		case "23503": // foreign_key_violation
			return NewAppError(400, ErrValidation, "invalid_request", err)
		}
	}

	return NewAppError(500, ErrInternal, "internal_error", err)
}

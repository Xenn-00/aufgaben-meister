package tx

import (
	"context"

	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
)

type Tx interface {
	Commit(ctx context.Context) *app_errors.AppError
	Rollback(ctx context.Context) *app_errors.AppError
}

type TxManager interface {
	Begin(ctx context.Context) (Tx, *app_errors.AppError)
}

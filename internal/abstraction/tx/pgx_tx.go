package tx

import (
	"context"

	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxTxManager struct {
	db *pgxpool.Pool
}

func NewPgxTxManager(db *pgxpool.Pool) *PgxTxManager {
	return &PgxTxManager{db: db}
}

func (m *PgxTxManager) Begin(ctx context.Context) (Tx, *app_errors.AppError) {
	tx, err := m.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return &PgxTx{Tx: tx}, nil
}

type PgxTx struct {
	Tx pgx.Tx
}

func (t *PgxTx) Commit(ctx context.Context) *app_errors.AppError {
	if err := t.Tx.Commit(ctx); err != nil {
		return app_errors.NewAppError(
			fiber.StatusInternalServerError,
			app_errors.ErrInternal,
			"internal_error",
			err,
		)
	}
	return nil
}

func (t *PgxTx) Rollback(ctx context.Context) *app_errors.AppError {
	_ = t.Tx.Rollback(ctx)
	return nil
}

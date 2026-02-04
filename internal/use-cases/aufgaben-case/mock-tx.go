package aufgaben_case

import (
	"context"

	"github.com/Xenn-00/aufgaben-meister/internal/abstraction/tx"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/stretchr/testify/mock"
)

type MockTx struct {
	mock.Mock
}

func (m *MockTx) Commit(ctx context.Context) *app_errors.AppError {
	args := m.Called(ctx)
	return args.Get(0).(*app_errors.AppError)
}

func (m *MockTx) Rollback(ctx context.Context) *app_errors.AppError {
	args := m.Called(ctx)
	return args.Get(0).(*app_errors.AppError)
}

type MockTxManager struct {
	mock.Mock
}

func (m *MockTxManager) Begin(ctx context.Context) (tx.Tx, *app_errors.AppError) {
	args := m.Called(ctx)
	return args.Get(0).(tx.Tx), args.Get(1).(*app_errors.AppError)
}

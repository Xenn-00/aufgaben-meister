package use_cases

import (
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/queue"
	worker_task "github.com/Xenn-00/aufgaben-meister/internal/worker/tasks"
	"github.com/stretchr/testify/mock"
)

var _ queue.TaskQueueClient = (*MockTaskQueue)(nil)

// Mock TaskQueue for testing
type MockTaskQueue struct {
	mock.Mock
}

func (m *MockTaskQueue) EnqueueSendInvitationEmail(payload *worker_task.SendInvitationEmailPayload) error {
	args := m.Called(payload)
	return args.Error(0)
}

func (m *MockTaskQueue) EnqueueSendProjectProgressReminder(payload *worker_task.SendProjectProgressReminder, scheduledAt time.Time) error {
	args := m.Called(payload, scheduledAt)
	return args.Error(0)
}

func (m *MockTaskQueue) EnqueueHandoverRequestNotifyMeister(payload *worker_task.HandoverRequestNotifyMeister) error {
	args := m.Called(payload)
	return args.Error(0)
}

package queue

import (
	"time"

	worker_task "github.com/Xenn-00/aufgaben-meister/internal/worker/tasks"
	"github.com/goccy/go-json"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type TaskQueue struct {
	client *asynq.Client
}

func NewTaskQueue(redis *redis.Client) *TaskQueue {
	return &TaskQueue{
		client: asynq.NewClientFromRedisClient(redis),
	}
}

func (q *TaskQueue) EnqueueSendInvitationEmail(payload *worker_task.SendInvitationEmailPayload) error {
	log.Info().Msg("Preparing enqueueing payload.")
	p, _ := json.Marshal(payload)
	task := asynq.NewTask(worker_task.TaskSendProjectInvitationEmail, p, asynq.Queue("email"), asynq.MaxRetry(5))

	_, err := q.client.Enqueue(task)
	return err
}

func (q *TaskQueue) EnqueueSendProjectProgressReminder(payload *worker_task.SendProjectProgressReminder, remindAt time.Time) error {
	log.Info().Msg("Preparing enqueueing payload.")
	p, _ := json.Marshal(payload)
	task := asynq.NewTask(worker_task.TaskSendProjectProgressReminder, p, asynq.Queue("low"), asynq.ProcessAt(remindAt))

	_, err := q.client.Enqueue(task)
	return err
}
func (q *TaskQueue) EnqueueHandoverRequestNotifyMeister(payload *worker_task.HandoverRequestNotifyMeister) error {
	log.Info().Msg("Preparing enqueueing payload.")
	p, _ := json.Marshal(payload)
	task := asynq.NewTask(worker_task.TaskHandoverRequestNotifyMeister, p, asynq.Queue("email"))

	_, err := q.client.Enqueue(task)
	return err
}

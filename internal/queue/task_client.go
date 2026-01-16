package queue

import (
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

func (q *TaskQueue) EnqueueSendInvitationEmail(invitationID, rawToken string) error {
	log.Info().Msg("Preparing enqueueing payload.")
	payload, _ := json.Marshal(worker_task.SendInvitationEmailPayload{
		InvitationID: invitationID,
		RawToken:     rawToken,
	})

	task := asynq.NewTask(worker_task.TaskSendProjectInvitationEmail, payload, asynq.Queue("email"), asynq.MaxRetry(5))

	_, err := q.client.Enqueue(task)
	return err
}

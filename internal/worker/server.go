package worker

import (
	worker_handler "github.com/Xenn-00/aufgaben-meister/internal/worker/handlers"
	worker_task "github.com/Xenn-00/aufgaben-meister/internal/worker/tasks"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func RunWorker(redis *redis.Client, handler *worker_handler.WorkerHander) error {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redis.Options().Addr,
			Password: redis.Options().Password,
			DB:       redis.Options().DB,
		},
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"email": 10,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(
		worker_task.TaskSendProjectInvitationEmail,
		handler.WorkerInvitationEmailHandler(),
	)

	return srv.Run(mux)
}

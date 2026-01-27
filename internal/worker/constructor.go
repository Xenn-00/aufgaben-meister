package worker

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func NewWorkerServer(redis *redis.Client) *asynq.Server {
	return asynq.NewServer(
		asynqRedisOpt(redis),
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"email":   6,
				"default": 3,
				"low":     1,
			},
			RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
				return time.Duration(n) * time.Second
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().
					Err(err).
					Str("task", task.Type()).
					Bytes("payload", task.Payload()).
					Msg("task failed")
			}),
		},
	)
}

func NewScheduler(redis *redis.Client) *asynq.Scheduler {
	return asynq.NewScheduler(
		asynqRedisOpt(redis),
		&asynq.SchedulerOpts{
			Location: time.Local,
			LogLevel: asynq.InfoLevel,
		},
	)
}

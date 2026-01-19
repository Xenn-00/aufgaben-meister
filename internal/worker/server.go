package worker

import (
	"context"
	"fmt"
	"time"

	worker_handler "github.com/Xenn-00/aufgaben-meister/internal/worker/handlers"
	worker_task "github.com/Xenn-00/aufgaben-meister/internal/worker/tasks"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func RunWorker(ctx context.Context, redis *redis.Client, handler *worker_handler.WorkerHander) error {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redis.Options().Addr,
			Password: redis.Options().Password,
			DB:       redis.Options().DB,
		},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"email":   6, // 60% workers (high priority)
				"default": 3, // 30% workers
				"low":     1, // 10% workers (background tasks)
			},
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return time.Duration(n) * time.Second // exponential backoff
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().
					Err(err).
					Str("task", task.Type()).
					Str("payload", string(task.Payload())).
					Msg("task failed")
			}),
		},
	)

	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr:     redis.Options().Addr,
			Password: redis.Options().Password,
			DB:       redis.Options().DB,
		},
		&asynq.SchedulerOpts{
			Location: time.Local,
			LogLevel: asynq.InfoLevel,
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(
		worker_task.TaskSendProjectInvitationEmail,
		handler.WorkerInvitationEmailHandler(),
	)
	mux.HandleFunc(worker_task.TaskInvitationExpire, handler.WorkerInvitationExpireHandler())

	_, err := scheduler.Register("0 0 * * *", asynq.NewTask(worker_task.TaskInvitationExpire, nil), asynq.Queue("low"))
	if err != nil {
		return fmt.Errorf("failed to register scheduler: %w", err)
	}

	log.Info().Msg("scheduled task: check expired invitations everyday at midnight.")

	// Run scheduler
	go func() {
		if err := scheduler.Run(); err != nil {
			log.Error().Err(err).Msg("scheduler error")
		}
	}()

	// Run worker
	go func() {
		if err := srv.Run(mux); err != nil {
			log.Error().Err(err).Msg("worker server error")
		}
	}()

	// wait for shutdown
	<-ctx.Done()
	log.Info().Msg("shutting down worker server...")

	scheduler.Shutdown()
	srv.Shutdown()

	return nil
}

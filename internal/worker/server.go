package worker

import (
	"context"

	worker_handler "github.com/Xenn-00/aufgaben-meister/internal/worker/handlers"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func RunWorker(ctx context.Context, redis *redis.Client, handler *worker_handler.WorkerHander) error {
	srv := NewWorkerServer(redis)
	scheduler := NewScheduler(redis)

	mux := asynq.NewServeMux()
	RegisterWorkerHandlers(mux, handler)

	if err := RegisterCronJobs(scheduler); err != nil {
		return err
	}

	go func() {
		if err := scheduler.Run(); err != nil {
			log.Error().Err(err).Msg("scheduler error")
		}
	}()

	go func() {
		if err := srv.Run(mux); err != nil {
			log.Error().Err(err).Msg("worker error")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutting down workers...")

	scheduler.Shutdown()
	srv.Shutdown()
	return nil
}

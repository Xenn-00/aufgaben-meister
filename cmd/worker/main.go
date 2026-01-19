package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/config"
	"github.com/Xenn-00/aufgaben-meister/internal/db"
	"github.com/Xenn-00/aufgaben-meister/internal/mail"
	"github.com/Xenn-00/aufgaben-meister/internal/worker"
	worker_handler "github.com/Xenn-00/aufgaben-meister/internal/worker/handlers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	cfg := config.LoadConfig()

	dbPool := db.ConnectPool(cfg.DATABASE.Postgres.DSN)
	redisPool, err := db.RedisPool(cfg.DATABASE.Redis.Addr, cfg.DATABASE.Redis.Password, 0)
	if err != nil {
		log.Fatal().Err(err)
	}

	mailer := mail.NewMailer(cfg)
	handler := worker_handler.NewWorkerHandler(dbPool, mailer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// run worker
	errChan := make(chan error, 1)
	go func() {
		log.Info().Msg("Starting worker server...")
		if err := worker.RunWorker(ctx, redisPool, handler); err != nil {
			errChan <- err
		}
	}()

	// wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		log.Info().Str("signal", sig.String()).Msg("received shutdown signal")
		cancel()
		dbPool.Close()
		redisPool.Close()
		log.Info().Msg("worker shutdown complete")
	case err := <-errChan:
		log.Fatal().Err(err).Msg("worker crashed")
	}
}

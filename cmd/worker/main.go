package main

import (
	"os"
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

	if err := worker.RunWorker(redisPool, handler); err != nil {
		log.Fatal().Err(err)
	}
}

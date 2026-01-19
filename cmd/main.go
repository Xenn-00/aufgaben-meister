package main

// Package main ist der Einstiegspunkt der Anwendung "aufgaben-meister".
// Es verantwortet das Laden der Konfiguration, die Initialisierung der
// Datenbankverbindung und des Paseto-Tokenmakers, das Aufsetzen der Fiber-API
// mit Middleware und Routern sowie das Starten des HTTP-Servers.

// Hinweise zur Weiterentwicklung / TODOs:
// - Tests für Graceful-Shutdown-Szenarien ergänzen (z. B. lang laufende Handler).
// - Konfigurationsvalidierung nach dem Laden (z. B. Port-Format, DSN-Prüfung).
// - Optionalen Healthcheck-Endpoint hinzufügen, um Readiness/Liveness zu signalisieren.

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/config"
	"github.com/Xenn-00/aufgaben-meister/internal/db"
	"github.com/Xenn-00/aufgaben-meister/internal/i18n"
	"github.com/Xenn-00/aufgaben-meister/internal/middleware"
	"github.com/Xenn-00/aufgaben-meister/internal/routers"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// main initialisiert alle benötigten Ressourcen für den HTTP-Server und stellt sicher,
// dass bei Beendigung sauber heruntergefahren und aufgeräumt wird.
func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	// 0. I18N Einführung
	i18nSvc := i18n.NewInitI18nService()
	// 1. Konfiguration laden (config.LoadConfig).
	cfg := config.LoadConfig()
	// 2. Postgres-Verbindungs-Pool (db.ConnectPool) und Redis-Verbindungs-Pool erstellen.
	dbPool := db.ConnectPool(cfg.DATABASE.Postgres.DSN)
	redisPool, err := db.RedisPool(cfg.DATABASE.Redis.Addr, cfg.DATABASE.Redis.Password, 0)
	if err != nil {
		log.Fatal().Err(err) // Fehler beim Initialisieren Redis-Pool
	}
	// 3. Paseto-Maker initialisieren (utils.NewPasetoMaker).
	paseto, err := utils.NewPasetoMaker(cfg.APP_SECRET.Paseto.HexKey) // Paseto erwartet einen gültigen HexKey in cfg.APP_SECRET.Paseto.HexKey.
	if err != nil {
		log.Fatal().Err(err) // Fehler beim Initialisieren (z. B. beim Paseto-Maker) führen zum sofortigen Abbruch.
	}

	// 4. Fiber-App mit ErrorHandler, RequestID- und Logger-Middleware erstellen.
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandlerMiddleware(i18nSvc),
	})
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.AcceptLanguageMiddleware())
	app.Use(middleware.LoggerMiddleware())

	// 5. Applikationsrouten registrieren (routers.SetupRoutes).
	cfgStorage := routers.CfgRedisStorage{
		Host:     cfg.DATABASE.Redis.Addr,
		Password: cfg.DATABASE.Redis.Password,
	}
	routers.SetupRoutes(app, dbPool, redisPool, i18nSvc, paseto, cfgStorage)

	go func() {
		// 6. HTTP-Server starten (app.Listen), soll es am Ende stellen, weil es blocking ist. Aber wir können es eigenlich in eine Goroutine einfügen.
		log.Info().Msgf("Starte %s auf Port %s", cfg.APP.Name, cfg.APP.Port) // Logging erfolgt mit zerolog
		if err := app.Listen(fmt.Sprintf(":%s", cfg.APP.Port)); err != nil { // Ports und App-Metadaten werden aus cfg.APP gelesen; falsche Konfiguration verhindert Start.
			if err == http.ErrServerClosed {
				log.Info().Msg("Server ordnungsgemäß herunterfahren.")
			} else {
				log.Fatal().Err(err).Msgf("Der Server konnte nicht gestartet werden, %v", err)
			}
		}
	}()

	// 7. Graceful Shutdown bei SIGINT/SIGTERM: DB-Pool schließen, Fiber herunterfahren,
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM) // Signale werden mit signal.NotifyContext abgefangen, stop() wird deferred aufgerufen.
	<-ctx.Done()
	stop()
	log.Warn().Msg("Shutdown-Signal empfangen... Vorbereitung zum Herunterfahren.")

	if redisPool != nil {
		redisPool.Close()
		log.Info().Msg("Redis-Pool erfolgreich geschlossen.")
	}

	if dbPool != nil { // Vor dem Schließen des DB-Pools auf nil prüfen (pool != nil), um Panics zu vermeiden.
		dbPool.Close()
		log.Info().Msg("DB-Pool erfolgreich geschlossen.")
	}
	log.Info().Msg("Datenbank-Pool erfolgreich abgeschlossen.")

	// Fiber shutdown
	if err := app.Shutdown(); err != nil { // app.Shutdown() beendet laufende Verbindungen und Handler sauber; Fehler dabei sollten geloggt werden.
		log.Error().Err(err).Msgf("Beim Herunterfahren ist ein Fehler aufgtreten: %v", err)
	}
	log.Info().Msg("Server ordnungsgemäß herunterfahren.")
}

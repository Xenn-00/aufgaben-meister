package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ConnectPool richtet einen Verbindungs-Pool zur Datenbank ein.
func ConnectPool(dsn string) *pgxpool.Pool {
	// Parsen der DSN
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Err(err).Msg("Fehler beim Parsen der Datenbank-DSN")
		return nil
	}

	// Erstellen des verbindungspools
	cfg.MaxConns = 20                       // Maximale Anzahl der Verbindungen im Pool
	cfg.MinConns = 5                        // Minimale Anzahl der Verbindungen im Pool
	cfg.MaxConnIdleTime = time.Hour         // Maximale Leerlaufzeit einer Verbindung
	cfg.HealthCheckPeriod = time.Minute * 5 // Periodische Überprüfung der Verbindungen

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Err(err).Msg("Fehler beim Erstellen des Datenbank-Pools")
		return nil
	}

	return pool
}

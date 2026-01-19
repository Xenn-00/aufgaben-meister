package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// RedisPool erstellt und konfigurierten Redis-Client und prüft die Erreichbarkeit.
// RedisPool nimmt die Verbindungsparameter addr (host:port), password (leer = keine Auth) und db (DB-Index) entgegen und initialisiert einen *redis.Client mit vordefinierten Pool- und Timeout-Einstellungen.
// Wichtige Hinweise:
// Der Caller ist verantwortlich dafür, den Client mit client.Close() zu schließen, wenn er nicht mehr benötigt wird.
// Es werden Standardwerte für PoolSize, MaxIdleConns sowie Dial/Read/Write-Timeouts (jeweils 5s) gesetzt — bei Bedarf anpassen.
// Ein erfolgreicher Rückgabewert bedeutet, dass die Verbindung zum Zeitpunkt des Aufrufs funktioniert; Laufzeitfehler beim späteren Gebrauch sind dennoch möglich.
func RedisPool(addr, password string, db int) (*redis.Client, error) {
	// *redis.Client: konfigurierter, goroutine-sicherer Redis-Client (verwendet Connection-Pooling).
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     20,
		MaxIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Die Funktion führt vor der Rückgabe ein PING mit einem 5‑Sekunden-Timeout aus und liefert einen Fehler, wenn die Verbindung nicht aufgebaut werden kann.
	if err := rdb.Ping(ctx).Err(); err != nil {
		// error: Fehler beim Verbindungsaufbau oder beim PING.
		log.Error().Err(err).Msg(fmt.Errorf("Fehler beim Erstellen des Redis-Pool: %w", err).Error())
		return nil, fmt.Errorf("Verbindung zu Redis nicht möglich: %w", err)
	}

	return rdb, nil
}

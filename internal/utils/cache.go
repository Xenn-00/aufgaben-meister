package utils

import (
	"context"
	"time"

	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	json "github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// GetCacheData versucht, einen Wert aus Redis zu lesen und in den generischen Typ T zu unmarshalen.
// Parameter: ctx für Request-Scoping, rdb Redis-Client, cacheKey der Schlüssel.
// Rückgabe: *T (Pointer auf deserialisiertes Objekt) und *app_errors.AppError (nil bei Erfolg).
// Hinweis: nutzt goccy/go-json; erwartet JSON als gespeicherten Wert.
func GetCacheData[T any](ctx context.Context, rdb *redis.Client, cacheKey string) (*T, *app_errors.AppError) {
	val, err := rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		// Gibt bei Cache-Miss (Key nicht vorhanden) (nil, nil) zurück.
		return nil, nil // Cache-miss
	} else if err != nil {
		// Bei Redis-Fehlern oder JSON-Unmarshal-Fehlern wird (nil, *app_errors.AppError) zurückgegeben.
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	var data T
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	return &data, nil
}

// SetCacheData serialisiert das gegebene Objekt (T) als JSON und speichert es mit Ablaufzeit in Redis.
// Parameter: ctx, rdb, cacheKey, data Pointer auf zu speicherndes Objekt, expire Dauer bis Ablauf.
// Hinweis: Daten werden als JSON-Bytes gespeichert.
func SetCacheData[T any](ctx context.Context, rdb *redis.Client, cacheKey string, data *T, expire time.Duration) *app_errors.AppError {
	bytes, err := json.Marshal(data)
	if err != nil {
		// Bei JSON-Marshal-Fehler wird ein *app_errors.AppError zurückgegeben.
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	// Sonst wird das Ergebnis von rdb.Set(...).Err() zurückgegeben (nil bei Erfolg, Fehler sonst).
	if err := rdb.Set(ctx, cacheKey, bytes, expire).Err(); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return nil
}

// DeleteCacheData löscht den angegebenen cacheKey aus Redis.
// Parameter: ctx, rdb, cacheKey.
// Hinweis: kein Fehler, wenn Key bereits nicht existiert.
func DeleteCacheData(ctx context.Context, rdb *redis.Client, cacheKey string) error {
	// Gibt das Ergebnis von rdb.Del(...).Err() zurück (nil bei Erfolg).
	return rdb.Del(ctx, cacheKey).Err()
}

package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthRouter registriert Health- und Readiness-Endpoints auf dem gegebenen Fiber-Router.
// Parameter:
//   - app: Ziel-Router (fiber.Router), auf dem die Routen registriert werden.
//   - db:  Postgres-Verbindungs-Pool (*pgxpool.Pool) für den Readiness-Check.
func HealthRouter(app fiber.Router, db *pgxpool.Pool, redis *redis.Client) {
	// Endpoints:
	//   - GET /healthz: Liefert eine JSON-Antwort mit Statusinformation (HTTP 200).
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "Health-OK",
			"message": "Service lebt.",
		})
	})
	//   - GET /livez:  Einfache Liveness-Antwort als Text (HTTP 200).
	app.Get("/livez", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Lebt.")
	})
	//   - GET /readyz: Führt einen Readiness-Check durch, indem die Datenbank mit db.Ping() geprüft wird.
	app.Get("/readyz", func(c *fiber.Ctx) error {
		// Überprüft des Verbindungs der Redis
		if err := redis.Ping(c.Context()); err != nil {
			//	Bei einem Verbindungsfehler wird HTTP 503 mit einer JSON-Fehlermeldung zurückgegeben,
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "Fehlversuch",
				"error":  "Redis ist nicht bereit.",
			})
		}

		// Überprüft des Verbindungs der Datenbank
		if err := db.Ping(c.Context()); err != nil {
			//	Bei einem Verbindungsfehler wird HTTP 503 mit einer JSON-Fehlermeldung zurückgegeben,
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "Fehlversuch",
				"error":  "Datenbank ist nicht bereit.",
			})
		}
		//	sonst HTTP 200 mit einer Bestätigung, dass Datenbank und App einsatzbereit sind.
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "Bereit",
			"message": "Datenbank und App sind einsatzbereit.",
		})
	})
}

package middleware

import (
	"fmt"
	"strings"

	"github.com/Xenn-00/aufgaben-meister/internal/dtos"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// AuthMiddleware validiert das Authorization-Header ("Bearer <token>") und verifiziert das PASETO-Token.
// pasetoMaker: *utils.PasetoMaker zum Verifizieren von PASETO-Tokens.
// Verhalten:
// - Sendet bei fehlendem Header, falschem Format oder ungültigem/abgelaufenem Token HTTP 401 mit einer JSON-Fehlerantwort.
// - Bei erfolgreicher Verifizierung setzt es die Context-Lokale: "user_id", "username", "email", "role".
// - Liefert einen fiber.Handler zurück, der bei Erfolg c.Next() aufruft.
func AuthMiddleware(pasetoMaker *utils.PasetoMaker, redis *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status": "error",
				"error": dtos.ErrorResponse{
					Code:    fiber.StatusUnauthorized,
					Message: "Authorization header fehlt",
				},
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status": "error",
				"error": dtos.ErrorResponse{
					Code:    fiber.StatusUnauthorized,
					Message: "Token-Format ist falsch. Nutze Bearer <token>.",
				},
			})
		}

		token := parts[1]

		// Verifizieren via PASETO
		payload, err := pasetoMaker.VerifyToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status": "error",
				"error": dtos.ErrorResponse{
					Code:    fiber.StatusUnauthorized,
					Message: "Token ist ungültig oder abgelaufen (1)", // 1 => Token kann nicht verifiziert werden
				},
			})
		}

		// Überprüft ein Token, ob es noch in Redis oder nicht ist.
		device := c.Get("X-Device-Name")
		if device == "" {
			device = "Unknown Device"
		}

		redisKey := fmt.Sprintf("user_sessions:%s:%s", payload.UserID, device)
		session, _ := redis.HGetAll(c.Context(), redisKey).Result()
		if len(session) == 0 || session["token"] != token {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status": "error",
				"error": dtos.ErrorResponse{
					Code:    fiber.StatusUnauthorized,
					Message: "Token ist ungültig oder abgelaufen (2)", // 2 => Token ist nicht mehr in Redis
				},
			})
		}

		// Speichern zu kontext, sodass Handler es nutzen kann
		c.Locals("user_id", payload.UserID)
		c.Locals("username", payload.Username)
		c.Locals("email", payload.Email)
		c.Locals("role", payload.Role)
		c.Locals("device_name", device)

		return c.Next()
	}
}

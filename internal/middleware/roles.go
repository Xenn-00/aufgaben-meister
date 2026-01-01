package middleware

import (
	"slices"

	"github.com/Xenn-00/aufgaben-meister/internal/dtos"
	"github.com/gofiber/fiber/v2"
)

// RequireRoles prüft, ob die im Context unter "role" gespeicherte Rolle einer der erlaubten Rollen (allowedRoles) entspricht.
func RequireRoles(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role == "" {
			// Ist keine Rolle vorhanden, wird ein 401 Unauthorized zurückgegeben. Bei fehlender Berechtigung erfolgt ein 403 Forbidden.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status": "error",
				"error": dtos.ErrorResponse{
					Code:    fiber.StatusUnauthorized,
					Message: "Kein Zugriff, keine Rolle gefunden.",
				},
			})
		}

		if slices.Contains(allowedRoles, role) {
			// Bei erfolgreicher Prüfung wird der nächste Handler (c.Next()) aufgerufen.
			return c.Next()
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status": "error",
			"error": dtos.ErrorResponse{
				Code:    fiber.StatusForbidden,
				Message: "Sie haben hier nicht zu melden",
			},
		})
	}
}

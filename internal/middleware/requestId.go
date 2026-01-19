package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// RequestIDMiddleware fügt jeder Anfrage eine eindeutige Anforderungs-ID hinzu.
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Überprüfe, ob die Anforderungs-ID bereits gesetzt ist
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			id, err := gonanoid.New()
			if err != nil {
				return fmt.Errorf("Fehler beim Generieren der Anforderungs-ID: %w", err)
			}
			requestID = fmt.Sprintf("AM-%s", id)
		}

		// Setze die Anforderungs-ID in den Kontext und die Antwortheader
		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

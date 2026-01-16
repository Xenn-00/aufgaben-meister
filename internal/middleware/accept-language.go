package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AcceptLanguageMiddleware ruft Accept-Language von Header ab und speichert den Wert bei c.Locals
func AcceptLanguageMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Get("Accept-Language", "en")
		lang := strings.Split(raw, ",")[0] //Denn in der Praxis hat der Header „Accept-Language“ immer das folgende Format: „de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7“.
		lang = strings.Split(lang, "-")[0]
		c.Locals("lang", lang)
		return c.Next()
	}
}

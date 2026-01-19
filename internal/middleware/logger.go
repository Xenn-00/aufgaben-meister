package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// LoggerMiddleware protokolliert eingehende Anfragen und deren Antworten.
func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Protokolliere die eingehende Anfrage
		start := time.Now()
		err := c.Next()
		method := c.Method()
		duration := time.Since(start)

		reqID := c.Locals("request_id").(string)

		log.Info().Str("[request_id]", reqID).Msgf("%s %s (%v) %d", method, c.Path(), duration, c.Response().StatusCode())

		return err
		// reqID, _ := c.Locals("request_id").(string)

		// logger := log.With().
		// 	Str("request_id", reqID).
		// 	Str("path", c.Path()).
		// 	Str("method", c.Method()).
		// 	Logger()

		// ctx := logger.WithContext(c.Context())
		// c.SetUserContext(ctx)

		// return c.Next()

	}
}

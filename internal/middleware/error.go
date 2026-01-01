package middleware

import "github.com/gofiber/fiber/v2"

// ErrorHandlerMiddleware behandelt Fehler, die während der Anfrageverarbeitung auftreten.
func ErrorHandlerMiddleware(c *fiber.Ctx, err error) error {
	// Überprüfe, ob der Fehler ein *fiber.Error ist
	if fiberErr, ok := err.(*fiber.Error); ok {
		// Sende die entsprechende Fehlerantwort
		return c.Status(fiberErr.Code).JSON(fiber.Map{
			"status":     "error",
			"message":    fiberErr.Message,
			"request_id": c.Locals("request_id").(string),
		})
	}

	code := fiber.StatusInternalServerError

	if e, ok := err.(interface{ Code() int }); ok {
		code = e.Code()
	}

	return c.Status(code).JSON(fiber.Map{
		"status":     "error",
		"message":    err.Error(),
		"request_id": c.Locals("request_id"),
	})
}

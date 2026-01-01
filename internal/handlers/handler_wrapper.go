package handlers

import (
	"github.com/Xenn-00/aufgaben-meister/internal/dtos"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/gofiber/fiber/v2"
)

type HandlerFunc func(c *fiber.Ctx) *app_errors.AppError

// HandlerWrapper/Wrap wickelt einen HandlerFunc ab und behandelt Fehler.
func Wrap(fn HandlerFunc) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := fn(c); err != nil {
			return c.Status(err.Code).JSON(fiber.Map{
				"status": "error",
				"error": dtos.ErrorResponse{
					Code:    err.Code,
					Message: err.Message,
					Field:   err.Field,
				},
			})
		}
		return nil
	}
}

// CreateResponse erstellt eine standardisierte WebResponse.
func CreateResponse[T any](message string, data T, requestID string, addition ...T) dtos.WebResponse[T] {
	return dtos.WebResponse[T]{
		Message:   message,
		Data:      data,
		RequestID: requestID,
		Addition:  addition,
	}
}

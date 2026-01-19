package handlers

import (
	"github.com/Xenn-00/aufgaben-meister/internal/dtos"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/gofiber/fiber/v2"
)

// CreateResponse erstellt eine standardisierte WebResponse.
func CreateResponse[T any](message string, data T, requestID string, details ...T) dtos.WebResponse[T] {
	return dtos.WebResponse[T]{
		Message:   message,
		Data:      data,
		RequestID: requestID,
		Details:   details,
	}
}

func GetUserID(c *fiber.Ctx) (string, *app_errors.AppError) {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return "", app_errors.NewAppError(fiber.StatusUnauthorized, app_errors.ErrUnauthorized, "auth.unauthorized", nil)
	}

	return userID, nil
}

func GetRequestID(c *fiber.Ctx) string {
	reqID, ok := c.Locals("request_id").(string)
	if !ok {
		reqID = "unknown"
	}
	return reqID
}

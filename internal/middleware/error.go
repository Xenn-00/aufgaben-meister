package middleware

import (
	"errors"

	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	internal_i18n "github.com/Xenn-00/aufgaben-meister/internal/i18n"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// ErrorHandlerMiddleware behandelt Fehler, die während der Anfrageverarbeitung auftreten.
func ErrorHandlerMiddleware(i18nSvc internal_i18n.Service) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		lang := c.Get("Accept-Language", "en")

		var appErr *app_errors.AppError
		if !errors.As(err, &appErr) {
			appErr = app_errors.NewAppError(
				fiber.StatusInternalServerError,
				app_errors.ErrInternal,
				"internal_error",
				err,
			)
		}

		message := i18nSvc.T(lang, appErr.MessageKey, nil)

		reqID, _ := c.Locals("request_id").(string)

		respErr := fiber.Map{
			"code":       appErr.Code,
			"type":       appErr.Type,
			"message":    message,
			"request_id": reqID,
		}

		if len(appErr.Details) > 0 {
			var details []fiber.Map

			for _, d := range appErr.Details {
				details = append(details, fiber.Map{
					"field":  d.Field,
					"reason": d.Reason,
					"message": i18nSvc.T(
						lang,
						d.MessageKey,
						d.Params,
					),
				})
			}

			respErr["details"] = details
		}

		if appErr.Err != nil {
			log.Error().Err(appErr.Err).Msg("application error")
		}

		return c.Status(appErr.Code).JSON(fiber.Map{
			"status": "error",
			"error":  respErr,
		})
	}
}

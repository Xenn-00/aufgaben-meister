package handlers

import (
	"strings"
	"unicode"

	"github.com/Xenn-00/aufgaben-meister/internal/dtos"
	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	project_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/project-dto"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// CreateResponse erstellt eine standardisierte WebResponse.
func CreateResponse[T any](message string, data T, requestID string, details ...any) dtos.WebResponse[T] {
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

func GetParamProjectID(c *fiber.Ctx, v *validator.Validate) (string, *app_errors.AppError) {
	var param project_dto.ParamProjectID
	if err := c.ParamsParser(&param); err != nil {
		return "", app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidParam, "request.invalid_param", err)
	}

	if err := v.Struct(param); err != nil {
		return "", app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}
	return param.ID, nil
}
func GetParamTaskID(c *fiber.Ctx, v *validator.Validate) (string, *app_errors.AppError) {
	var param aufgaben_dto.ParamTaskID
	if err := c.ParamsParser(&param); err != nil {
		return "", app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidParam, "request.invalid_param", err)
	}

	if err := v.Struct(param); err != nil {
		return "", app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}
	return param.ID, nil
}

func NormalizeStatusCase(s string) string {
	// Lowercase first
	s = strings.ToLower(s)

	// Replace with underscore
	s = strings.ReplaceAll(s, " ", "_")

	// Title case every word (after underscore)
	words := strings.Split(s, "_")
	for i, word := range words {
		if len(word) > 0 {
			// Capitalize first letter only
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}

	return strings.Join(words, "_")
}

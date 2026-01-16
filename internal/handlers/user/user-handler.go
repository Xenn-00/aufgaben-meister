package user_handlers

import (
	user_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/user-dto"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/Xenn-00/aufgaben-meister/internal/handlers"
	internal_i18n "github.com/Xenn-00/aufgaben-meister/internal/i18n"
	user_case "github.com/Xenn-00/aufgaben-meister/internal/use-cases/user-case"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type UserHandler struct {
	validator *validator.Validate
	service   user_case.UserServiceContract
	i18n      internal_i18n.Service
}

// Geschützt mit AuthMiddleware
func NewUserHandler(db *pgxpool.Pool, redis *redis.Client, i18n *internal_i18n.I18nService) *UserHandler {
	validate := validator.New()
	return &UserHandler{
		validator: validate,
		service:   user_case.NewUserService(db, redis),
		i18n:      i18n,
	}
}

func (h *UserHandler) FetchUserSelfProfile(c *fiber.Ctx) error {
	// Wir benötigen keine Anfrage um Benutzerprofils zu kriegen
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)

	resp, err := h.service.UserSelfProfile(c.Context(), userID)
	if err != nil {
		return err
	}

	lang := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_get_self", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}
	return nil
}

func (h *UserHandler) FetchUserProfile(c *fiber.Ctx) error {
	// Wir brauchen nur UserID von c.Locals
	// Wir kriegen das suchtende Benutzer-ID von Parameter
	var req user_dto.ParamGetUserByID

	if err := c.ParamsParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidParam, "request.invalid_param", err)
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)

	resp, err := h.service.UserProfileById(c.Context(), req, userID)
	if err != nil {
		return err
	}

	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_fetch_user", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *UserHandler) UpdateSelfProfile(c *fiber.Ctx) error {
	// 0. Krieg die Benutzer-ID aus c.Locals
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	var req user_dto.UpdateSelfProfileRequest
	// 1. Anfrage parsen
	if err := c.BodyParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", err)
	}

	// 2. Überprüft die Anfrage, ob alle Felder in der Anfrage leere String sind oder nicht. Wenn alle Felder in der Anfrage leere String sind, ablehnen die einfach.
	if req.Username == "" && req.Email == "" && req.Name == "" {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.body_empty", nil)
	}

	// 3. Validieren
	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// 4. Service aufrufen
	resp, err := h.service.UpdateSelfProfile(c.Context(), req, userID)
	if err != nil {
		return err
	}

	// 5. Krieg die Request ID
	reqID := handlers.GetRequestID(c)

	// 6. Antwort zurückgeben
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_update_self", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *UserHandler) DeactivateSelfUser(c *fiber.Ctx) error {
	var req user_dto.DeactivateSelfUserRequest

	if err := c.BodyParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", err)
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	if err := h.service.DeactivateSelfUser(c.Context(), req, userID); err != nil {
		return nil
	}

	reqID := handlers.GetRequestID(c)

	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_self_deactivate", nil), "Ok", reqID)
	if err := c.Status(fiber.StatusNoContent).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

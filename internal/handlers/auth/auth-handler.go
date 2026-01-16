package auth_handlers

import (
	auth_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/auth-dto"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/Xenn-00/aufgaben-meister/internal/handlers"
	internal_i18n "github.com/Xenn-00/aufgaben-meister/internal/i18n"
	auth_case "github.com/Xenn-00/aufgaben-meister/internal/use-cases/auth-case"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AuthHandler struct {
	validator *validator.Validate
	service   auth_case.AuthServiceContract
	i18n      internal_i18n.Service
}

func NewAuthHandler(db *pgxpool.Pool, redis *redis.Client, i18n *internal_i18n.I18nService, paseto *utils.PasetoMaker) *AuthHandler {
	validate := validator.New()
	return &AuthHandler{
		validator: validate,
		i18n:      i18n,
		service:   auth_case.NewAuthService(db, redis, paseto),
	}
}

// RegisterUser behandelt die Registrierung eines neuen Benutzers.
func (h *AuthHandler) RegisterUser(c *fiber.Ctx) error {
	// TODO: Was soll hier passieren?
	// 1. Anfrage parsen
	var req auth_dto.RegisterUserRequest

	if err := c.BodyParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", err)
	}
	// 2. Validieren
	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}
	// 3. Service aufrufen
	resp, err := h.service.RegisterUser(c.Context(), req)
	if err != nil {
		return err
	}

	// 4. Krieg die Request ID
	reqID := handlers.GetRequestID(c)

	// 5. Antwort zurückgeben

	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_register", nil), resp, reqID)
	if err := c.Status(fiber.StatusCreated).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}
	return nil
}

// LoginUser behandelt die Anmeldung eines Benutzers.
func (h *AuthHandler) LoginUser(c *fiber.Ctx) error {
	// TODO: Was soll hier passieren?
	// 1. Anfrage parsen
	var req auth_dto.LoginUserRequest

	if err := c.BodyParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrValidation, "request.invalid_body", err)
	}
	// 2. Validieren
	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// 3. Login metadata erstellen
	ua := c.Get("User-Agent")
	if ua == "" {
		ua = "Unknown-Test-Client"
	}

	device := c.Get("X-Device-Name")
	device = detectDeviceType(ua)

	loginMetadata := auth_dto.LoginMetadata{
		UserAgent: ua,
		Device:    device,
		IP:        c.IP(),
	}

	// 3. Service aufrufen
	resp, err := h.service.LoginUser(c.Context(), req, loginMetadata)
	if err != nil {
		return err
	}

	// 4. Krieg die Request ID
	reqID := handlers.GetRequestID(c)

	// 5. Antwort zurückgeben
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_login", nil), resp, reqID)
	if err := c.Status(fiber.StatusCreated).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}
	return nil

}

// LogoutUser beendet die Sitzung eines authentifizierten Benutzers für ein bestimmtes Gerät.
func (h *AuthHandler) LogoutUser(c *fiber.Ctx) error {
	// LogoutUser braucht keine Anfrage
	// Wir benötigen stattdessen eine JTI von c.Locals
	jti, ok := c.Locals("jti").(string)
	if !ok || jti == "" {
		return app_errors.NewAppError(fiber.StatusUnauthorized, app_errors.ErrUnauthorized, "auth.unauthorized", nil)
	}

	if err := h.service.LogoutUser(c.Context(), jti); err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)

	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "respons.success_logout", nil), "OK", reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}
	return nil
}

// ListAllUserDevices listet alle registrierten Geräte eines Benutzers auf.
// Hinweis:
//   - Erwartet, dass eine vorgelagerte Middleware die Authentifizierung übernimmt
//     und "user_id" in c.Locals setzt.
func (h *AuthHandler) ListAllUserDevices(c *fiber.Ctx) error {
	// Erwartete c.Locals:
	//   - "user_id" (string): ID des authentifizierten Benutzers (erforderlich)
	//   - "request_id" (string): optional, wird zur Antwort-Trace verwendet
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)

	// Ruft h.service.ListAllUserDevices(ctx, userID) auf, um die Geräteliste zu erhalten.
	devices, err := h.service.ListAllUserDevices(c.Context(), userID)
	if err != nil {
		// Gibt den vom Service zurückgegebenen Fehler unverändert weiter.
		return err
	}

	resp := map[string]any{
		"devices": devices,
	}

	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_list_device", nil), resp, reqID, map[string]any{"count_devices": len(*devices)})
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		// Gibt 500 Internal Server Error, falls die Antwort nicht serialisiert/gesendet werden kann.
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *AuthHandler) LogoutAllDevices(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)

	if err := h.service.LogoutAllDevices(c.Context(), userID); err != nil {
		return err
	}

	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_logout_all", nil), "OK", reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		// Gibt 500 Internal Server Error, falls die Antwort nicht serialisiert/gesendet werden kann.
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

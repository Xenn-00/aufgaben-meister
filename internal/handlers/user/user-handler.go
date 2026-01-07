package user_handler

import (
	"fmt"

	user_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/user-dto"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/Xenn-00/aufgaben-meister/internal/handlers"
	user_case "github.com/Xenn-00/aufgaben-meister/internal/use-cases/user-case"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type UserHandler struct {
	validator *validator.Validate
	service   user_case.UserServiceContract
}

// Geschützt mit AuthMiddleware
func NewUserHandler(db *pgxpool.Pool, redis *redis.Client) *UserHandler {
	validate := validator.New()
	return &UserHandler{
		validator: validate,
		service:   user_case.NewUserService(db, redis),
	}
}

func (h *UserHandler) FetchUserSelfProfile(c *fiber.Ctx) *app_errors.AppError {
	// Wir benötigen keine Anfrage um Benutzerprofils zu kriegen
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		// Gibt 401 Unauthorized zurück, wenn "user_id" fehlt oder leer ist.
		return app_errors.New(fiber.StatusUnauthorized, "Nicht authorisiert. Kein Benutzer gefunden", "Nicht-Authorisiert")
	}

	reqID, ok := c.Locals("request_id").(string)
	if !ok {
		reqID = "unknown"
	}

	resp, err := h.service.UserSelfProfile(c.Context(), userID)
	if err != nil {
		return err
	}

	webResp := handlers.CreateResponse("Erfolgreiches Abrufen des Benutzerprofils", resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.New(fiber.StatusInternalServerError, fmt.Sprintf("Fehler beim Senden der Antwort: %v", err), "Antwort-Fehler")
	}
	return nil
}

func (h *UserHandler) FetchUserProfile(c *fiber.Ctx) *app_errors.AppError {
	// Wir brauchen nur UserID von c.Locals
	// Wir kriegen das suchtende Benutzer-ID von Parameter
	var req user_dto.ParamGetUserByID

	if err := c.ParamsParser(&req); err != nil {
		return app_errors.New(fiber.StatusBadRequest, fmt.Sprintf("Der Parameter kann nicht geparst werden: %v", err), "Ungültige-Parameter")
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.New(fiber.StatusBadRequest, fmt.Sprintf("Der parameter ist ungültig: %v", err), "Ungültige-Parameter")
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		// Gibt 401 Unauthorized zurück, wenn "user_id" fehlt oder leer ist.
		return app_errors.New(fiber.StatusUnauthorized, "Nicht authorisiert. Kein Benutzer gefunden", "Nicht-Authorisiert")
	}

	reqID, ok := c.Locals("request_id").(string)
	if !ok {
		reqID = "unknown"
	}

	resp, err := h.service.UserProfileById(c.Context(), req, userID)
	if err != nil {
		return err
	}
	webResp := handlers.CreateResponse("Erfolgreiches Abrufen des Benutzerprofils", resp, reqID)
	if err := c.Status(fiber.StatusCreated).JSON(webResp); err != nil {
		return app_errors.New(fiber.StatusInternalServerError, fmt.Sprintf("Fehler beim Senden der Antwort: %v", err), "Antwort-Fehler")
	}

	return nil
}

func (h *UserHandler) UpdateSelfProfile(c *fiber.Ctx) *app_errors.AppError {
	// 0. Krieg die Benutzer-ID aus c.Locals
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		// Gibt 401 Unauthorized zurück, wenn "user_id" fehlt oder leer ist.
		return app_errors.New(fiber.StatusUnauthorized, "Nicht authorisiert. Kein Benutzer gefunden", "Nicht-Authorisiert")
	}

	var req user_dto.UpdateSelfProfileRequest
	// 1. Anfrage parsen
	if err := c.BodyParser(&req); err != nil {
		return app_errors.New(fiber.StatusBadRequest, fmt.Sprintf("Die Anfrage kann nicht geparst werden: %v", err), "Ungültige-Anfrage")
	}

	// 2. Überprüft die Anfrage, ob alle Felder in der Anfrage leere String sind oder nicht. Wenn alle Felder in der Anfrage leere String sind, ablehnen die einfach.
	if req.Username == "" && req.Email == "" && req.Name == "" {
		return app_errors.New(fiber.StatusBadRequest, "Die Anfrage kann nicht alles leer sein", "Ungültige-Anfrage")
	}

	// 3. Validieren
	if err := h.validator.Struct(req); err != nil {
		return app_errors.New(fiber.StatusBadRequest, fmt.Sprintf("Die Anfrage ist ungültig: %v", err), "Ungültige-Anfrage")
	}

	// 4. Service aufrufen
	resp, err := h.service.UpdateSelfProfile(c.Context(), req, userID)
	if err != nil {
		return err
	}

	// 5. Krieg die Request ID
	reqID, ok := c.Locals("request_id").(string)
	if !ok {
		reqID = "unknown"
	}

	// 6. Antwort zurückgeben
	webResp := handlers.CreateResponse("Benutzer-Profil erfolgreich aktualisiert.", resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.New(fiber.StatusInternalServerError, fmt.Sprintf("Fehler beim Senden der Antwort: %v", err), "Antwort-Fehler")
	}

	return nil
}

func (h *UserHandler) DeactivateSelfUser(c *fiber.Ctx) *app_errors.AppError {
	var req user_dto.DeactivateSelfUserRequest

	if err := c.BodyParser(&req); err != nil {
		return app_errors.New(fiber.StatusBadRequest, fmt.Sprintf("Die Anfrage kann nicht geparst werden: %v", err), "Ungültige-Anfrage")
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.New(fiber.StatusBadRequest, fmt.Sprintf("Die Anfrage ist ungültig: %v", err), "Ungültige-Anfrage")
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		// Gibt 401 Unauthorized zurück, wenn "user_id" fehlt oder leer ist.
		return app_errors.New(fiber.StatusUnauthorized, "Nicht authorisiert. Kein Benutzer gefunden", "Nicht-Authorisiert")
	}

	if err := h.service.DeactivateSelfUser(c.Context(), req, userID); err != nil {
		return nil
	}

	reqID, ok := c.Locals("request_id").(string)
	if !ok {
		reqID = "unknown"
	}

	webResp := handlers.CreateResponse("Benutzer-Profil erfolgreich Deaktiviert.", "Ok", reqID)
	if err := c.Status(fiber.StatusNoContent).JSON(webResp); err != nil {
		return app_errors.New(fiber.StatusInternalServerError, fmt.Sprintf("Fehler beim Senden der Antwort: %v", err), "Antwort-Fehler")
	}

	return nil
}

package auth_handlers

import (
	"fmt"

	auth_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/auth-dto"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/Xenn-00/aufgaben-meister/internal/handlers"
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
}

func NewAuthHandler(db *pgxpool.Pool, redis *redis.Client, paseto *utils.PasetoMaker) *AuthHandler {
	validate := validator.New()
	return &AuthHandler{
		validator: validate,
		service:   auth_case.NewAuthService(db, redis, paseto),
	}
}

// RegisterUser behandelt die Registrierung eines neuen Benutzers.
func (h *AuthHandler) RegisterUser(c *fiber.Ctx) *app_errors.AppError {
	// TODO: Was soll hier passieren?
	// 1. Anfrage parsen
	var req auth_dto.RegisterUserRequest

	if err := c.BodyParser(&req); err != nil {
		return app_errors.New(fiber.StatusBadRequest, fmt.Sprintf("Die Anfrage kann nicht analysiert werden: %v", err), "Ungültige-Anfrage")
	}
	// 2. Validieren
	if err := h.validator.Struct(req); err != nil {
		return app_errors.New(fiber.StatusBadRequest, fmt.Sprintf("Die Anfrage ist ungültig: %v", err), "Ungültige-Anfrage")
	}
	// 3. Service aufrufen
	resp, err := h.service.RegisterUser(c.Context(), req)
	if err != nil {
		return err
	}

	// 4. Krieg die Request ID
	reqID, ok := c.Locals("request_id").(string)
	if !ok {
		reqID = "unknown"
	}

	// 5. Antwort zurückgeben
	webResp := handlers.CreateResponse("Benützer erfolgreich registriert", resp, reqID)
	if err := c.Status(fiber.StatusCreated).JSON(webResp); err != nil {
		return app_errors.New(fiber.StatusInternalServerError, fmt.Sprintf("Fehler beim Senden der Antwort: %v", err), "Antwort-Fehler")
	}
	return nil
}

// LoginUser behandelt die Anmeldung eines Benutzers.
func (h *AuthHandler) LoginUser(c *fiber.Ctx) *app_errors.AppError {
	// TODO: Was soll hier passieren?
	// 1. Anfrage parsen
	var req auth_dto.LoginUserRequest

	if err := c.BodyParser(&req); err != nil {
		return app_errors.New(fiber.StatusBadRequest, fmt.Sprintf("Die Anfrage kann nicht analysiert werden: %v", err), "Ungültige-Anfrage")
	}
	// 2. Validieren
	if err := h.validator.Struct(req); err != nil {
		return app_errors.New(fiber.StatusBadRequest, fmt.Sprintf("Die Anfrage ist ungültig: %v", err), "Ungültige-Anfrage")
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
	reqID, ok := c.Locals("request_id").(string)
	if !ok {
		reqID = "unknown"
	}

	// 5. Antwort zurückgeben
	webResp := handlers.CreateResponse("Benützer erfolgreich anmelden", resp, reqID)
	if err := c.Status(fiber.StatusCreated).JSON(webResp); err != nil {
		return app_errors.New(fiber.StatusInternalServerError, fmt.Sprintf("Fehler beim Senden der Antwort: %v", err), "Antwort-Fehler")
	}
	return nil

}

// LogoutUser beendet die Sitzung eines authentifizierten Benutzers für ein bestimmtes Gerät.
func (h *AuthHandler) LogoutUser(c *fiber.Ctx) *app_errors.AppError {
	// LogoutUser braucht keine Anfrage
	// Wir benötigen stattdessen eine Benutzer-ID von c.Locals
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return app_errors.New(fiber.StatusUnauthorized, "Nicht authorisiert. Kein Benutzer gefunden", "Kein-Benutzer")
	}

	deviceName, ok := c.Locals("device_name").(string)
	if !ok || deviceName == "" {
		return app_errors.New(fiber.StatusUnauthorized, "Gerät kann nicht erkannt werden", "Unbekanntes-Gerät")
	}

	if err := h.service.LogoutUser(c.Context(), userID, deviceName); err != nil {
		return err
	}

	reqID, ok := c.Locals("request_id").(string)
	if !ok {
		reqID = "unknown"
	}

	webResp := handlers.CreateResponse("Benutzer erfolgreich abmelden", "OK", reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.New(fiber.StatusInternalServerError, fmt.Sprintf("Fehler beim Senden der Antwort: %v", err), "Antwort-Fehler")
	}
	return nil
}

// ListAllUserDevices listet alle registrierten Geräte eines Benutzers auf.
//
// Beschreibung:
//
//	Liest die Benutzer-ID aus c.Locals("user_id") und ruft den Auth-Service auf,
//	um alle zugehörigen Geräte zu holen. Die erfolgreiche Antwort wird als JSON
//	mit HTTP-Status 200 zurückgegeben.
//
// Hinweis:
//   - Erwartet, dass eine vorgelagerte Middleware die Authentifizierung übernimmt
//     und "user_id" in c.Locals setzt.
func (h *AuthHandler) ListAllUserDevices(c *fiber.Ctx) *app_errors.AppError {
	// Erwartete c.Locals:
	//   - "user_id" (string): ID des authentifizierten Benutzers (erforderlich)
	//   - "request_id" (string): optional, wird zur Antwort-Trace verwendet
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		// Gibt 401 Unauthorized zurück, wenn "user_id" fehlt oder leer ist.
		return app_errors.New(fiber.StatusUnauthorized, "Nicht authorisiert. Kein Benutzer gefunden", "Kein-Benutzer")
	}

	reqID, ok := c.Locals("request_id").(string)
	if !ok {
		reqID = "unknown"
	}

	// Ruft h.service.ListAllUserDevices(ctx, userID) auf, um die Geräteliste zu erhalten.
	devices, err := h.service.ListAllUserDevices(c.Context(), userID)
	if err != nil {
		// Gibt den vom Service zurückgegebenen Fehler unverändert weiter.
		return err
	}

	resp := map[string]any{
		"devices": devices,
	}

	webResp := handlers.CreateResponse("Alle Geräte des Benutzers erfolgreich auflisten.", resp, reqID, map[string]any{"count_devices": len(*devices)})
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		// Gibt 500 Internal Server Error, falls die Antwort nicht serialisiert/gesendet werden kann.
		return app_errors.New(fiber.StatusInternalServerError, fmt.Sprintf("Fehler beim Senden der Antwort: %v", err), "Antwort-Fehler")
	}

	return nil
}

func (h *AuthHandler) LogoutAllDevices(c *fiber.Ctx) *app_errors.AppError {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		// Gibt 401 Unauthorized zurück, wenn "user_id" fehlt oder leer ist.
		return app_errors.New(fiber.StatusUnauthorized, "Nicht authorisiert. Kein Benutzer gefunden", "Kein-Benutzer")
	}

	reqID, ok := c.Locals("request_id").(string)
	if !ok {
		reqID = "unknown"
	}

	if err := h.service.LogoutAllDevices(c.Context(), userID); err != nil {
		return err
	}

	webResp := handlers.CreateResponse("Alle Geräte wurden erfolgreich abgemeldet.", "OK", reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		// Gibt 500 Internal Server Error, falls die Antwort nicht serialisiert/gesendet werden kann.
		return app_errors.New(fiber.StatusInternalServerError, fmt.Sprintf("Fehler beim Senden der Antwort: %v", err), "Antwort-Fehler")
	}

	return nil
}

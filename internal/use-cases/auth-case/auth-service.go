package auth_case

import (
	"context"
	"fmt"
	"regexp"
	"time"

	auth_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/auth-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	auth_repo "github.com/Xenn-00/aufgaben-meister/internal/repo/auth-repo"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type AuthService struct {
	// Hier können Abhängigkeiten wie Repositories
	// oder andere Services eingefügt werden.
	db     *pgxpool.Pool
	redis  *redis.Client
	paseto *utils.PasetoMaker
	repo   auth_repo.AuthRepoContract
}

func NewAuthService(db *pgxpool.Pool, redis *redis.Client, paseto *utils.PasetoMaker) AuthServiceContract {
	return &AuthService{
		repo:   auth_repo.NewAuthRepo(db),
		redis:  redis,
		paseto: paseto,
	}
}

// RegisterUser registriert einen neuen Benutzer.
func (s *AuthService) RegisterUser(ctx context.Context, req auth_dto.RegisterUserRequest) (*auth_dto.RegisterUserResponse, *app_errors.AppError) {
	// Überprüfen, ob der Benützer bereits existiert oder nicht.
	// Wenn nicht, dann neuen Benutzer erstellen.

	filter := &entity.UserCountFilter{
		Email:    &req.Email,
		Username: &req.Username,
	}

	count, err := s.repo.CountUsers(ctx, *filter)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, app_errors.New(fiber.StatusConflict, "Ein Benutzer mit dieser E-Mail oder diesem Benutzernamen existiert bereits.", "Benutzer-Existiert")
	}

	// Passwort hashen
	hashed, hashErr := utils.GenerateHash(req.Password)
	if hashErr != nil {
		return nil, app_errors.New(fiber.StatusInternalServerError, hashErr.Error(), "Password-Hashen")
	}

	// Neuen Benutzer erstellen
	idUser, idErr := uuid.NewV7()
	if idErr != nil {
		return nil, app_errors.New(fiber.StatusInternalServerError, idErr.Error(), "ID-Generiert")
	}

	newUser := &entity.UserEntity{
		ID:           idUser.String(),
		Email:        req.Email,
		Name:         req.Name,
		Username:     req.Username,
		PasswordHash: hashed,
	}

	newUserID, err := s.repo.SaveUsers(ctx, *newUser)
	if err != nil {
		return nil, err
	}

	// Token erstellen
	token, pasetoErr := s.paseto.CreateToken(newUserID, newUser.Username, newUser.Email, true, 15*time.Minute)
	if pasetoErr != nil {
		return nil, app_errors.New(fiber.StatusInternalServerError, pasetoErr.Error(), "Paseto-Generiert")
	}

	// Response zurücksenden
	response := &auth_dto.RegisterUserResponse{
		UserID: newUserID,
		Token:  token,
	}

	return response, nil
}

// LoginUser authentifiziert einen Benutzer anhand von E-Mail oder Benutzername,
// validiert das Passwort, erzeugt ein Paseto-Token (TTL: 15 Minuten) und legt
// eine Sitzungsinformation in Redis ab.
//
// Parameter:
//   - ctx: Kontext für Request-Lifetime und Cancellation.
//   - req: LoginUserRequest mit Identifier (E-Mail oder Benutzername) und Passwort.
//   - loginMeta: Metadaten zur Sitzung (Device, UserAgent, IP). Falls Device leer ist,
//     wird "Unknown Device" verwendet.
//
// Rückgaben:
//   - Bei Erfolg: *auth_dto.LoginUserResponse mit UserID und Token.
//   - Bei Fehler: *app_errors.AppError (z. B. Benutzer nicht gefunden, ungültiges Passwort,
//     Token-Generierung, Redis-Fehler).
//
// Nebeneffekte:
//   - Speichert eine Sitzung unter dem Schlüssel "user_sessions:{userID}:{device}" mit
//     Ablaufzeit (15 Minuten).
//   - Loggt interne Suchfehler beim Finden des Benutzers.
//
// Hinweise:
// - Der Identifier wird per Regex auf E-Mail geprüft; andernfalls als Benutzername verwendet.
// - Passwortprüfung erfolgt mit utils.VerifyHash.
func (s *AuthService) LoginUser(ctx context.Context, req auth_dto.LoginUserRequest, loginMeta auth_dto.LoginMetadata) (*auth_dto.LoginUserResponse, *app_errors.AppError) {
	identifier := req.Identifier

	// Anmelden per E-Mail oder Benutzername
	emailRegex := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	var user *entity.UserEntity
	var err *app_errors.AppError

	if emailRegex.MatchString(identifier) {
		user, err = s.repo.FindByEmail(ctx, identifier)
	} else {
		user, err = s.repo.FindByUsername(ctx, identifier)
	}

	if err != nil {
		log.Error().Err(err).Msgf("Fehler beim Suchen der Benutzer: %v", err)
		return nil, app_errors.New(fiber.StatusUnauthorized, "Falsches Anmeldedaten.", "Ungültige-Anmelden")
	}

	redisKey := fmt.Sprintf("user_sessions:%s:%s", user.ID, loginMeta.Device)
	sessionCached, cachedErr := utils.GetCacheData[SessionTracker](ctx, s.redis, redisKey)
	if sessionCached != nil {
		response := &auth_dto.LoginUserResponse{
			UserID: user.ID,
			Token:  sessionCached.Token,
		}

		return response, nil
	}

	if cachedErr != nil {
		return nil, cachedErr
	}

	// Passwort úberprüfen
	if isValid, err := utils.VerifyHash(user.PasswordHash, req.Password); !isValid || err != nil {
		return nil, app_errors.New(fiber.StatusUnauthorized, "Falsches Anmeldedaten.", "Ungültige-Anmelden")
	}

	// Überprüft der Benutzer, ob der noch Aktiv ist oder nicht
	if user.IsActive == false {
		// Wieder Aktiviert
		tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return nil, app_errors.New(fiber.StatusInternalServerError, "Fehler bei der Initialisierung der DB-Transaktion", "DB-Transaktion-Fehler")
		}
		defer tx.Rollback(ctx)
		if _, activateErr := s.repo.UserActivate(ctx, tx, user.ID); activateErr != nil {
			return nil, activateErr
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, app_errors.New(fiber.StatusInternalServerError, "Fehler beim Commit der DB-Transaktion", "DB-Transaktion-Fehler")
		}
	}

	// Token erstellen
	token, pasetoErr := s.paseto.CreateToken(user.ID, user.Username, user.Email, user.IsActive, 15*time.Minute)
	if pasetoErr != nil {
		return nil, app_errors.New(fiber.StatusInternalServerError, pasetoErr.Error(), "Paseto-Generiert")
	}

	if loginMeta.Device == "" {
		loginMeta.Device = "Unknown Device"
	}

	session := &SessionTracker{
		Token:    token,
		Device:   loginMeta.Device,
		UserAgen: loginMeta.UserAgent,
		IP:       loginMeta.IP,
		LoginAt:  time.Now().Format(time.RFC3339),
	}
	utils.SetCacheData(ctx, s.redis, redisKey, session, 15*time.Minute)

	// Response zurücksenden
	response := &auth_dto.LoginUserResponse{
		UserID: user.ID,
		Token:  token,
	}

	return response, nil
}

// LogoutUser beendet die Sitzung eines Benutzers für ein bestimmtes Gerät.
// Wenn deviceName leer ist, wird "Unknown Device" verwendet. Die Methode
// entfernt den zugehörigen Redis-Eintrag (Schlüssel: "user_sessions:{userID}:{device}").
// Rückgabe: nil bei Erfolg, ansonsten ein *app_errors.AppError bei Fehlern.
func (s *AuthService) LogoutUser(ctx context.Context, userID string, deviceName string) *app_errors.AppError {
	if deviceName == "" {
		deviceName = "Unknown Device"
	}

	redisKey := fmt.Sprintf("user_sessions:%s:%s", userID, deviceName)
	if err := utils.DeleteCacheData(ctx, s.redis, redisKey); err != nil {
		return app_errors.New(fiber.StatusInternalServerError, "Fehler beim Löschen von Redis-Daten", "Redis-Löschung")
	}

	return nil
}

// ListAllUserDevices ruft alle aktiven Geräte/Sessions eines Benutzers aus Redis ab.
func (s *AuthService) ListAllUserDevices(ctx context.Context, userID string) (*[]auth_dto.ListAllUserDevicesResponse, *app_errors.AppError) {
	pattern := fmt.Sprintf("user_sessions:%s:*", userID)
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()

	var devices []auth_dto.ListAllUserDevicesResponse

	for iter.Next(ctx) {
		key := iter.Val()
		data, err := utils.GetCacheData[SessionTracker](ctx, s.redis, key)
		if err != nil {
			return nil, app_errors.New(fiber.StatusInternalServerError, "Fehler beim Abrufen der Geräte.", "Redis-Scan")
		}

		var loginAt time.Time
		if ts := data.LoginAt; ts != "" {
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				loginAt = t
			} else {
				log.Error().Err(err).Msgf("Ungültiges login_at-Format für Schlüssel %s: %v", key, err)
			}
		}

		devices = append(devices, auth_dto.ListAllUserDevicesResponse{
			Key:       key,
			Device:    data.Device,
			IP:        data.IP,
			UserAgent: data.UserAgen,
			LoginAt:   loginAt,
		})
	}

	if err := iter.Err(); err != nil {
		return nil, app_errors.New(fiber.StatusInternalServerError, "Scan Fehler", "redis-scan")
	}

	return &devices, nil
}

// LogoutAllDevices löscht alle aktiven Sessions/Geräte eines Benutzers aus Redis.
func (s *AuthService) LogoutAllDevices(ctx context.Context, userID string) *app_errors.AppError {
	pattern := fmt.Sprintf("user_sessions:%s:*", userID)

	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := utils.DeleteCacheData(ctx, s.redis, iter.Val()); err != nil {
			return app_errors.New(fiber.StatusInternalServerError, "Fehler beim Löschen der Sitzung.", "Logout-All")
		}
	}

	if err := iter.Err(); err != nil {
		return app_errors.New(fiber.StatusInternalServerError, "Redis Scan Fehler.", "logout-scan")
	}

	return nil
}

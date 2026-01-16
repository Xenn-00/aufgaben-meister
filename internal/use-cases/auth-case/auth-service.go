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
		log.Debug().Msg("User already exist")
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict", nil)
	}

	// Passwort hashen
	hashed, hashErr := utils.GenerateHash(req.Password)
	if hashErr != nil {
		log.Error().Err(hashErr).Msg("An Error occured when trying to generate password hash")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", hashErr)
	}

	// Neuen Benutzer erstellen
	idUser, idErr := uuid.NewV7()
	if idErr != nil {
		log.Error().Err(idErr).Msg("An Error occured when trying to generate uuid v7")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", idErr)
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
	sessionID, sessionErr := uuid.NewV7()
	if sessionErr != nil {
		log.Error().Err(sessionErr).Msg("An Error occured when trying to generate uuid v7")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", sessionErr)
	}
	token, pasetoErr := s.paseto.CreateToken(newUserID, newUser.Username, newUser.Email, sessionID.String(), true, 15*time.Minute)
	if pasetoErr != nil {
		log.Error().Err(pasetoErr).Msg("An Error occured when trying to generate paseto token")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", pasetoErr)
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
		log.Error().Err(err.Err).Msgf("Fehler beim Suchen der Benutzer: %v", err)
		return nil, app_errors.NewAppError(fiber.StatusUnauthorized, app_errors.ErrUnauthorized, "auth.unauthorized", err.Err)
	}

	// Passwort úberprüfen
	if isValid, err := utils.VerifyHash(user.PasswordHash, req.Password); !isValid || err != nil {
		log.Error().Err(err).Msg("Fehler beim Passwort Verifiziert")
		return nil, app_errors.NewAppError(fiber.StatusUnauthorized, app_errors.ErrUnauthorized, "auth.unauthorized", err)
	}

	// Überprüft der Benutzer, ob der noch Aktiv ist oder nicht
	if !user.IsActive {
		// Wieder Aktiviert
		tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			log.Error().Err(err).Msg("Fehler beim Starten der DB-Transaktion")
			return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
		}
		defer tx.Rollback(ctx)
		if _, activateErr := s.repo.UserActivate(ctx, tx, user.ID); activateErr != nil {
			return nil, activateErr
		}
		if err := tx.Commit(ctx); err != nil {
			log.Error().Err(err).Msg("Fehler beim Ausführen der DB-Transaktion")
			return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
		}
	}

	// SessionID erstellen
	sessionID, sessionError := uuid.NewV7()
	if sessionError != nil {
		log.Err(sessionError).Msg("An Error occured when trying to generate uuid v7")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", sessionError)
	}

	// Token erstellen
	token, pasetoErr := s.paseto.CreateToken(user.ID, user.Username, user.Email, sessionID.String(), user.IsActive, 15*time.Minute)
	if pasetoErr != nil {
		log.Error().Err(pasetoErr).Msg("Fehler beim Erstellen der Paseto-Token")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", pasetoErr)
	}

	if loginMeta.Device == "" {
		loginMeta.Device = "Unknown Device"
	}

	redisKey := fmt.Sprintf("session:%s", sessionID)
	session := &SessionTracker{
		JTI:      sessionID.String(),
		UserID:   user.ID,
		Token:    token,
		Device:   loginMeta.Device,
		UserAgen: loginMeta.UserAgent,
		IP:       loginMeta.IP,
		LoginAt:  time.Now().Format(time.RFC3339),
	}
	utils.SetCacheData(ctx, s.redis, redisKey, session, 15*time.Minute)

	sessionListKey := fmt.Sprintf("user_sessions:%s", user.ID)
	s.redis.SAdd(ctx, sessionListKey, session.JTI)

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
func (s *AuthService) LogoutUser(ctx context.Context, sessionID string) *app_errors.AppError {
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	// Der Session Abgerufen
	session, err := utils.GetCacheData[SessionTracker](ctx, s.redis, sessionKey)
	if err != nil || session == nil {
		// Session bereits beendet / ungültig
		return app_errors.NewAppError(fiber.StatusUnauthorized, app_errors.ErrUnauthorized, "auth.unauthorized", nil)
	}

	userSessionKey := fmt.Sprintf("user_sessions:%s", session.UserID)

	// 1. Löschen der Haupt-Session
	if err := utils.DeleteCacheData(ctx, s.redis, sessionKey); err != nil {
		log.Error().Err(err).Msg("Fehler beim Löschen der Cache")
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	// 2. Löschen der JTI von dem Set der User-Sessions
	if err := s.redis.SRem(ctx, userSessionKey, session.JTI).Err(); err != nil {
		log.Error().Err(err).Msg("Fehler beim Löschen der Cache")
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return nil
}

// ListAllUserDevices ruft alle aktiven Geräte/Sessions eines Benutzers aus Redis ab.
func (s *AuthService) ListAllUserDevices(ctx context.Context, userID string) (*[]auth_dto.ListAllUserDevicesResponse, *app_errors.AppError) {
	key := fmt.Sprintf("user_sessions:%s", userID)
	jtis, err := s.redis.SMembers(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Msg("Fehler beim Abrufen der Redis-SMembers")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	var devices []auth_dto.ListAllUserDevicesResponse
	for _, jti := range jtis {
		sessionKey := fmt.Sprintf("session:%s", jti)
		data, err := utils.GetCacheData[SessionTracker](ctx, s.redis, sessionKey)
		if err != nil {
			log.Error().Err(err.Err).Msg("Fehler beim Abrufen von Redis-Cache")
			return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err.Err)
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

	return &devices, nil
}

// LogoutAllDevices löscht alle aktiven Sessions/Geräte eines Benutzers aus Redis.
func (s *AuthService) LogoutAllDevices(ctx context.Context, userID string) *app_errors.AppError {
	key := fmt.Sprintf("user_sessions:%s:*", userID)
	jtis, err := s.redis.SMembers(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Msg("Fehler beim Abrufen der Redis-SMembers")
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	for _, jti := range jtis {
		sessionKey := fmt.Sprintf("session:%s", jti)
		if err := utils.DeleteCacheData(ctx, s.redis, sessionKey); err != nil {
			log.Error().Err(err).Msg("Fehler beim Löschen der Cache")
			return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
		}
	}

	if err := utils.DeleteCacheData(ctx, s.redis, key); err != nil {
		log.Error().Err(err).Msg("Fehler beim Löschen der Cache")
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return nil
}

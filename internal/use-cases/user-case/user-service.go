package user_case

import (
	"context"
	"fmt"
	"time"

	user_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/user-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	user_repo "github.com/Xenn-00/aufgaben-meister/internal/repo/user-repo"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type UserService struct {
	redis *redis.Client
	db    *pgxpool.Pool
	repo  user_repo.UserRepoContract
}

func NewUserService(db *pgxpool.Pool, redis *redis.Client) UserServiceContract {
	return &UserService{
		redis: redis,
		db:    db,
		repo:  user_repo.NewUserRepo(db),
	}
}

func (s *UserService) UserSelfProfile(ctx context.Context, userID string) (*user_dto.UserProfileResponse, *app_errors.AppError) {
	// 1. Verusuch, die Daten aus Redis zu laden
	// Redis dient nur als Cache, NICHT als Source of Truth
	redisKey := fmt.Sprintf("user_profile:%s", userID)
	cachedData, cachedErr := utils.GetCacheData[user_dto.UserProfileResponse](ctx, s.redis, redisKey)
	if cachedData != nil && cachedErr == nil {
		return cachedData, nil
	}

	// Bei Cache-Fehlern wird bewusst fortgefahren (Fallback auf DB)
	user, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Antwort basierend auf Rolle und Berechtigung vorbereiten
	resp := &user_dto.UserProfileResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.CreatedAt,
	}

	// Antwort im Cache speichern
	if err := utils.SetCacheData(ctx, s.redis, redisKey, resp, 15*time.Minute); err != nil {
		log.Error().Err(err.Err).Msg("Fehler beim Einstellen der Redis-Cache")
		return nil, err
	}

	return resp, nil
}

func (s *UserService) UserProfileById(ctx context.Context, req user_dto.ParamGetUserByID, viewerUserID string) (*user_dto.UserProfileResponse, *app_errors.AppError) {
	// 1. Verusuch, die Daten aus Redis zu laden
	// Redis dient nur als Cache, NICHT als Source of Truth
	redisKey := fmt.Sprintf("GET:userID=%s:viewerID=%s", req.ID, viewerUserID)
	cachedData, cachedErr := utils.GetCacheData[user_dto.UserProfileResponse](ctx, s.redis, redisKey)
	if cachedData != nil && cachedErr == nil {
		return cachedData, nil
	}
	// Bei Cache-Fehlern wird bewusst fortgefahren (Fallback auf DB)

	// 2. Autorisierungsprüfung
	// Überprüfen, ob beide benutzer mindestens einem gemainsamen Project angehören
	isUnderOneProject, err := s.repo.IsUnderOneProject(ctx, viewerUserID, req.ID)
	if err != nil {
		return nil, err
	}

	// 3. Benutzer inklusive Projekte aus der Datenbank laden
	user, err := s.repo.FindUserWithProjects(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 4. Antwort basierend auf Rolle und Berechtigung vorbereiten
	var resp *user_dto.UserProfileResponse

	if isUnderOneProject {
		// Volle Sicht für Benutzer innerhalb desselben Projekts
		resp = &user_dto.UserProfileResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Project:  user.Projects,
		}
	} else {
		// Eingeschränkte Sicht für andere
		resp = &user_dto.UserProfileResponse{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
		}
	}

	// 5. Antwort im Cache speichern
	if err := utils.SetCacheData(ctx, s.redis, redisKey, resp, 15*time.Minute); err != nil {
		log.Error().Err(err.Err).Msg("Fehler beim Einstellen der Redis-Cache")
		return nil, err
	}

	return resp, nil
}

func (s *UserService) UpdateSelfProfile(ctx context.Context, req user_dto.UpdateSelfProfileRequest, userID string) (*user_dto.UserProfileResponse, *app_errors.AppError) {

	model := entity.UserUpdate{
		ID:        userID,
		Email:     req.Email,
		Username:  req.Username,
		Name:      req.Name,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Fehler beim Starten der DB-Transaktion")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer tx.Rollback(ctx)

	user, updateErr := s.repo.UpdateSelfProfileTx(ctx, tx, userID, model)
	if updateErr != nil {
		return nil, updateErr
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("Fehler beim Ausführen der DB-Transaktion")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	// Prepare response
	resp := &user_dto.UserProfileResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// revoke cache
	redisKey := fmt.Sprintf("user_profile:%s", userID)
	if err := utils.DeleteCacheData(ctx, s.redis, redisKey); err != nil {
		log.Error().Err(err).Msg("Fehler beim Löschen der Cache")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	// set new cache data after updated
	if err := utils.SetCacheData(ctx, s.redis, redisKey, resp, 15*time.Minute); err != nil {
		log.Error().Err(err.Err).Msg("Fehler beim Einstellen der Redis-Cache")
		return nil, err
	}

	return resp, nil
}

func (s *UserService) DeactivateSelfUser(ctx context.Context, req user_dto.DeactivateSelfUserRequest, userID string) *app_errors.AppError {
	user, userErr := s.repo.FindByUserID(ctx, userID)
	if userErr != nil {
		return userErr
	}

	if isValid, hashErr := utils.VerifyHash(user.PasswordHash, req.Password); !isValid || hashErr != nil {
		log.Error().Err(hashErr).Msg("Fehler beim Passwort Verifiziert")
		return app_errors.NewAppError(fiber.StatusUnauthorized, app_errors.ErrUnauthorized, "auth.unauthorized", hashErr)
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Fehler beim Starten der DB-Transaktion")
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer tx.Rollback(ctx)

	if isDeactivated, err := s.repo.DeactivateSelfUser(ctx, tx, userID); err != nil && isDeactivated == false {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("Fehler beim Ausführen der DB-Transaktion")
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

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

	return nil
}

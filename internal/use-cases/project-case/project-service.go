package project_case

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	project_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/project-dto"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/Xenn-00/aufgaben-meister/internal/queue"
	project_repo "github.com/Xenn-00/aufgaben-meister/internal/repo/project-repo"
	worker_task "github.com/Xenn-00/aufgaben-meister/internal/worker/tasks"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type ProjectService struct {
	redis     *redis.Client
	db        *pgxpool.Pool
	repo      project_repo.ProjectRepoContract
	taskQueue *queue.TaskQueue
}

func NewProjectService(db *pgxpool.Pool, redis *redis.Client) ProjectServiceContract {
	return &ProjectService{
		redis:     redis,
		db:        db,
		repo:      project_repo.NewUserRepo(db),
		taskQueue: queue.NewTaskQueue(redis),
	}
}

func (s *ProjectService) CreateNewProject(ctx context.Context, req project_dto.CreateNewProjectRequest, userID string) (*project_dto.CreateNewProjectResponse, *app_errors.AppError) {
	// Ground Principal
	// 1. 1 Project = 1 Meister (for now)
	// 2. Source of the truth role = project_members
	// 3. projects.master_id = convenience / denormalized
	// 4. All write = Transaction

	// Generate project id
	projectID, uuidErr := uuid.NewV7()
	if uuidErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", uuidErr)
	}

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Fehler beim Starten der DB-Transaction")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Create model project
	modelProject := &entity.ProjectEntity{
		ID:         projectID.String(),
		Name:       req.Name,
		Type:       entity.ProjectType(req.TypeProject),
		Visibility: entity.ProjectVisibility(req.Visibility),
		MasterID:   userID,
	}

	// Create project
	project, respErr := s.repo.InsertNewProject(ctx, tx, modelProject)
	if respErr != nil {
		return nil, respErr
	}

	// Add to project member as role master
	if err := s.repo.InsertNewProjectMember(ctx, tx, project.ID, userID, entity.MEISTER); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("Fehler beim Ausführen der DB-Transaktion")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	committed = true

	// Prepare response
	resp := &project_dto.CreateNewProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		TypeProject: string(project.Type),
		Visibility:  string(project.Visibility),
		MasterID:    project.MasterID,
	}

	return resp, nil
}

func (s *ProjectService) GetSelfProject(ctx context.Context, userID string) ([]*project_dto.SelfProjectResponse, *app_errors.AppError) {
	// Call repo
	projects, err := s.repo.GetSelfProject(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := make([]*project_dto.SelfProjectResponse, 0, len(projects))
	for _, p := range projects {
		resp = append(resp, &project_dto.SelfProjectResponse{
			ID:          p.ID,
			Name:        p.Name,
			TypeProject: string(p.Type),
			Visibility:  string(p.Visibility),
			MasterID:    p.MasterID,
			Role:        p.Role,
		})
	}
	return resp, nil
}

func (s *ProjectService) GetProjectDetail(ctx context.Context, projectID, userID string) (*project_dto.GetProjectDetailResponse, *app_errors.AppError) {
	// Get role
	role, err := s.repo.GetUserRoleInProject(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	// Get project
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var resp *project_dto.GetProjectDetailResponse

	resp = &project_dto.GetProjectDetailResponse{
		ID:          project.ID,
		Name:        project.Name,
		TypeProject: string(project.Type),
		Visibility:  string(project.Visibility),
		Role:        &role,
	}

	if role == string(entity.MEISTER) {
		res, err := s.repo.GetProjectMember(ctx, projectID)
		if err != nil {
			return nil, err
		}
		resp.MasterID = &project.MasterID
		resp.Members = res
	}

	return resp, nil
}

func (s *ProjectService) InviteProjectMember(ctx context.Context, projectID, userID string, req *project_dto.InviteProjectMemberRequest) (*project_dto.InviteProjectMemberResponse, *app_errors.AppError) {

	// check if project really exist
	if _, err := s.repo.IsProjectExist(ctx, projectID); err != nil {
		log.Error().Err(err).Msg("Fehler beim Abrufen der Repository")
		return nil, err
	}
	// check user role, must be meister role to perform this procedure
	userRole, err := s.repo.GetUserRoleInProject(ctx, userID, projectID)
	if err != nil {
		log.Error().Err(err).Msg("Fehler beim Abrufen der Repository")
		return nil, nil
	}

	if userRole != string(entity.MEISTER) {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}

	// preload data
	memberSet, err := s.repo.GetProjectMemberUserIDs(ctx, projectID)
	if err != nil {
		log.Error().Err(err).Msg("Fehler beim Abrufen der Repository")
		return nil, err
	}
	inviteSet, err := s.repo.GetPendingInvitations(ctx, projectID) // prevent double invite
	if err != nil {
		log.Error().Err(err).Msg("Fehler beim Abrufen der Repository")
		return nil, err
	}

	userExistSet, err := s.repo.GetUsersByIds(ctx, req.UserIDs)
	if err != nil {
		log.Error().Err(err).Msg("Fehler beim Abrufen der Repository")
		return nil, err
	}

	// classify users
	toInvite := make([]string, 0, len(req.UserIDs))
	skipped := make([]project_dto.SkippedUser, 0)

	for _, uid := range req.UserIDs {
		// user didn't exist
		if !userExistSet[uid] {
			skipped = append(skipped, project_dto.SkippedUser{
				UserID: uid,
				Reason: "user_not_found",
			})
			continue
		}

		// already being member
		if memberSet[uid] {
			skipped = append(skipped, project_dto.SkippedUser{
				UserID: uid,
				Reason: "invitation.already_member",
			})
			continue
		}

		// already invited but still in pending
		if inviteSet[uid] {
			skipped = append(skipped, project_dto.SkippedUser{
				UserID: uid,
				Reason: "invitation.already_invited",
			})
			continue
		}

		toInvite = append(toInvite, uid)
	}
	if len(toInvite) == 0 {
		return &project_dto.InviteProjectMemberResponse{
			Invited:      []string{},
			SkippedUsers: skipped,
		}, nil
	}

	// begin transaction
	tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", txErr)
	}
	defer tx.Rollback(ctx)

	var invs []entity.ProjectInvitationEntity
	var payloadTasks []worker_task.SendInvitationEmailPayload
	for _, uid := range toInvite {
		rawToken, err := gonanoid.New()
		if err != nil {
			return nil, app_errors.NewAppError(
				fiber.StatusInternalServerError,
				app_errors.ErrInternal,
				"token_generation_failed",
				err,
			)
		}
		tokenHash := sha256.Sum256([]byte(rawToken))
		inv := entity.ProjectInvitationEntity{
			ID:            uuid.NewString(),
			ProjectID:     projectID,
			InvitedUserID: uid,
			InvitedBy:     userID,
			Role:          entity.MITARBEITER,
			Status:        entity.PENDING,
			TokenHash:     hex.EncodeToString(tokenHash[:]),
			ExpiresAt:     time.Now().Add(7 * 24 * time.Hour), // Eine Woche
		}
		invs = append(invs, inv)
		payloadTasks = append(payloadTasks, worker_task.SendInvitationEmailPayload{
			InvitationID: inv.ID,
			RawToken:     rawToken,
		})
	}

	// insert batch
	if err := s.repo.BatchInsertProjectInvitation(ctx, tx, invs); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	// commit
	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("Fehler beim Ausführen der DB-Transaktion")
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	// worker send email
	for _, task := range payloadTasks {
		if err := s.taskQueue.EnqueueSendInvitationEmail(task.InvitationID, task.RawToken); err != nil {
			log.Error().Err(err).Msg("Fehler beim Stellen die Aufgabe in die Warteschalnge")
		}
	}

	return &project_dto.InviteProjectMemberResponse{
		Invited:      toInvite,
		SkippedUsers: skipped,
	}, nil
}

func (s *ProjectService) AcceptInvitationProject(ctx context.Context, req *project_dto.InvitationQueryRequest, userID string) (*project_dto.InvitationMemberAccepted, *app_errors.AppError) {
	// TODO:
	// 1. Get invitation details from db
	inv, err := s.repo.GetInvitationProjectByID(ctx, req.InvitationID)
	if err != nil {
		return nil, err
	}
	// 2. Verify token
	tokenHash := sha256.Sum256([]byte(req.Token))
	incoming := hex.EncodeToString(tokenHash[:])

	if incoming != inv.TokenHash {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}
	// 3. Verify userID
	if userID != inv.InvitedUserID {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}
	// + Before update user state, we should check expiry time from that invitation
	if time.Now().After(inv.ExpiresAt) {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}

	// 4. transactional calls: update user state in project_invitations, add user to project_member respect to project_id, set token_hash in project_invitations = NULL
	tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", txErr)
	}
	defer tx.Rollback(ctx)

	// + Update user state in project_inviations, set token_hash to NULL (so that client can't abuse this endpoint), and set accepted_at
	if err := s.repo.UpdateUserInvitationState(ctx, tx, inv.ID, string(entity.ACCEPTED)); err != nil {
		return nil, err
	}

	// add invited user as project member
	if err := s.repo.InsertNewProjectMember(ctx, tx, inv.ProjectID, userID, inv.Role); err != nil {
		return nil, err
	}

	// commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	// 5. preparing for response
	project, err := s.repo.GetProjectByID(ctx, inv.ProjectID)
	if err != nil {
		return nil, err
	}

	resp := &project_dto.InvitationMemberAccepted{
		ID:         project.ID,
		Name:       project.Name,
		Role:       string(inv.Role),
		AcceptedAt: time.Now(),
	}

	return resp, nil
}

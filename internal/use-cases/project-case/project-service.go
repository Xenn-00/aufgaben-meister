package project_case

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/dtos"
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
		payload := &worker_task.SendInvitationEmailPayload{
			InvitationID: task.InvitationID,
			RawToken:     task.RawToken,
		}
		if err := s.taskQueue.EnqueueSendInvitationEmail(payload); err != nil {
			log.Error().Err(err).Msg("Fehler beim Stellen die Aufgabe in die Warteschlange")
		}
	}

	return &project_dto.InviteProjectMemberResponse{
		Invited:      toInvite,
		SkippedUsers: skipped,
	}, nil
}

func (s *ProjectService) AcceptInvitationProject(ctx context.Context, req *project_dto.InvitationQueryRequest, userID string) (*project_dto.InvitationMemberAccepted, *app_errors.AppError) {
	// TODO:
	// 0. Start DB transaktion
	tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", txErr)
	}
	defer tx.Rollback(ctx)
	// 1. Get invitation details from db
	inv, err := s.repo.GetInvitationProjectByIDWithTx(ctx, tx, req.InvitationID)
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
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict", nil)
	}

	// 4. transactional calls: update user state in project_invitations, add user to project_member respect to project_id, set token_hash in project_invitations = NULL
	// + Update user state in project_inviations, set token_hash to NULL (so that client can't abuse this endpoint), and set accepted_at
	if err := s.repo.AcceptUserInvitationState(ctx, tx, inv.ID, string(entity.ACCEPTED)); err != nil {
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

func (s *ProjectService) GetSelfInvitationPending(ctx context.Context, userID string) ([]*project_dto.SelfProjectInvitationResponse, *app_errors.AppError) {
	// TODO:
	// We need to query all user's pending invitation no matter what role they play in existing project, because not always user gonna check their email.
	// it's gonna (maybe) helpful for FE to have user's "pending invitations" data

	invs, err := s.repo.GetUserPendingInvitations(ctx, userID)
	if err != nil {
		return nil, err
	}

	var resp []*project_dto.SelfProjectInvitationResponse

	for _, inv := range invs {
		resp = append(resp, &project_dto.SelfProjectInvitationResponse{
			ID:          inv.ID,
			ProjectID:   inv.ProjectID,
			ProjectName: *inv.ProjectName,
			Role:        string(inv.Role),
			Status:      string(inv.Status),
			ExpiresAt:   inv.ExpiresAt,
			InvitedBy:   inv.InvitedBy,
		})
	}

	return resp, nil
}

func (s *ProjectService) RejectProjectInvitation(ctx context.Context, invitationID, userID string) (*project_dto.RejectProjectInvitationResponse, *app_errors.AppError) {
	// TODO:
	// 0. Transaction begin
	tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", txErr)
	}
	defer tx.Rollback(ctx)
	// 1. Check if project invitation is valid
	inv, err := s.repo.GetInvitationProjectByIDWithTx(ctx, tx, invitationID)
	if err != nil {
		return nil, err
	}
	// 2. Check if invited_user_id == userID
	if inv.InvitedUserID != userID {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}
	// 3. Check Expires
	if time.Now().After(inv.ExpiresAt) {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict", nil)
	}
	// 4. Check if invitation status still in "Pending"
	if inv.Status != entity.PENDING {
		return nil, app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict", nil)
	}
	// 5. Perform update project invitation status
	if err := s.repo.RejectUserInvitationState(ctx, tx, invitationID, string(entity.REJECTED)); err != nil {
		return nil, nil
	}
	// 6. Commit
	if err := tx.Commit(ctx); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	// 7. Prepare response
	var resp *project_dto.RejectProjectInvitationResponse
	projectInv, err := s.repo.GetInvitationProjectByID(ctx, invitationID)
	if err != nil {
		return nil, err
	}

	resp = &project_dto.RejectProjectInvitationResponse{
		ID:        projectInv.ID,
		ProjectID: projectInv.ProjectID,
		Status:    string(projectInv.Status),
		ExpiresAt: projectInv.ExpiresAt,
	}

	return resp, nil
}

func (s *ProjectService) RevokeProjectInvitations(ctx context.Context, projectID, userID string, req *project_dto.RevokeProjectMemberRequest) (*project_dto.RevokeProjectMemberResponse, *app_errors.AppError) {
	// TODO
	// 0. Make sure someone who perform this action has a project role as 'Meister'
	// There are 2 type of conditions for revoke:
	// A. When invited user hasn't yet accepted the invitation (status == 'Pending')
	// B. When invited user has already accepted the invitation (status == 'Accepted')
	// Revoke strategy A:
	// 0. Start transaction to prevent double action from client
	// 1. Prepare for batching update (scope: project_invitation table)
	// 2. Commit
	// Revoke strategy B:
	// 0. Start transaction to prevent double action from client
	// 1. Prepare for updating their state in project_members by performing soft delete
	// 2. Prepare for updating their status in project_invitation
	// 3. Commit
	userRole, err := s.repo.GetUserRoleInProject(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	if userRole != string(entity.MEISTER) {
		return nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}

	tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", nil)
	}
	defer tx.Rollback(ctx)

	resp := &project_dto.RevokeProjectMemberResponse{
		Revoked:      []string{},
		RevokedUsers: []project_dto.RevokedUser{},
	}

	// A. Revoke Pending invitations
	pendingRevoked, err := s.repo.RevokePendingInvitations(ctx, tx, projectID, req.UserIDs)
	if err != nil {
		return nil, err
	}

	for _, uid := range pendingRevoked {
		resp.Revoked = append(resp.Revoked, uid)
		resp.RevokedUsers = append(resp.RevokedUsers, project_dto.RevokedUser{
			UserID: uid,
			Reason: "invitation_revoked_before_acceptance",
		})
	}

	// B. Revoke Accepted members
	acceptedRevoked, err := s.repo.RevokeAcceptedMembers(ctx, tx, projectID, req.UserIDs)
	if err != nil {
		return nil, err
	}

	for _, uid := range acceptedRevoked {
		resp.Revoked = append(resp.Revoked, uid)
		resp.RevokedUsers = append(resp.RevokedUsers, project_dto.RevokedUser{
			UserID: uid,
			Reason: "membership_revoked_after_acceptance",
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return resp, nil
}

func (s *ProjectService) ResendProjectInvitations(ctx context.Context, invitationID, userID string) *app_errors.AppError {
	// TODO
	// Check valid invitationID
	inv, err := s.repo.GetInvitationProjectByID(ctx, invitationID)
	if err != nil {
		return err
	}
	// Check Authority
	userRole, err := s.repo.GetUserRoleInProject(ctx, userID, inv.ProjectID)
	if err != nil {
		return err
	}

	if userRole != string(entity.MEISTER) {
		return app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}
	// Check if invitation still valid or not
	if time.Now().After(inv.ExpiresAt) {
		return app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict", nil)
	}

	if inv.Status != entity.PENDING {
		return app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict", nil)
	}
	// Rotate token (transactional)
	newToken, tokenErr := gonanoid.New()
	if tokenErr != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	tokenHash := sha256.Sum256([]byte(newToken))
	tx, txErr := s.db.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.RotateTokenInvitation(ctx, tx, invitationID, hex.EncodeToString(tokenHash[:]), time.Now().Add(7*24*time.Hour)); err != nil {
		return err
	}
	// Commit
	if err := tx.Commit(ctx); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	// Send to worker
	payload := &worker_task.SendInvitationEmailPayload{
		InvitationID: invitationID,
		RawToken:     newToken,
	}
	if err := s.taskQueue.EnqueueSendInvitationEmail(payload); err != nil {
		log.Error().Err(err).Msg("Fehler beim Stellen die Aufgabe in die Warteschlange")
	}

	return nil
}

func (s *ProjectService) GetInvitationsInProject(ctx context.Context, projectID, userID string, filters project_dto.FilterProjectInvitation) ([]*project_dto.InvitationsInProjectResponse, *dtos.CursorPaginationMeta, *app_errors.AppError) {
	// TODO
	// Check authority
	role, err := s.repo.GetUserRoleInProject(ctx, userID, projectID)
	if err != nil {
		return nil, nil, err
	}

	if role != string(entity.MEISTER) {
		return nil, nil, app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
	}

	// Get project info first
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}

	// Rule check: filters.Expired only available if filters.Status == "Pending"
	if filters.Expired != nil {
		if filters.Status == nil || strings.ToTitle(*filters.Status) != "Pending" {
			return nil, nil, app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidQuery, "invitation.expired_only_valid_for_pending", nil)
		}
	}

	// Validate pagination, if limit == 0, then set default limit to 10 or if limit > 100 then set limit to 100
	if filters.Limit == 0 {
		filters.Limit = 20
	} else if filters.Limit > 100 {
		filters.Limit = 100
	}

	// call repo
	rows, err := s.repo.ListInvitations(ctx, projectID, &filters)
	if err != nil {
		return nil, nil, err
	}

	// build response cursor
	hasMore := false
	if len(rows) > filters.Limit {
		hasMore = true
		rows = rows[:filters.Limit]
	}

	var nextCursor *time.Time
	if hasMore {
		nextCursor = &rows[len(rows)-1].CreatedAt
	}

	var data []*project_dto.InvitationsInProjectResponse
	for _, row := range rows {
		data = append(data, &project_dto.InvitationsInProjectResponse{
			InvitationID: row.ID,
			UserID:       row.InvitedUserID,
			ProjectName:  project.Name,
			Status:       string(row.Status),
			ExpiresAt:    row.ExpiresAt,
			CreatedAt:    row.CreatedAt,
		})
	}

	cursor := &dtos.CursorPaginationMeta{
		Limit:      filters.Limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}

	return data, cursor, nil
}

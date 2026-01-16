package project_repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) ProjectRepoContract {
	return &ProjectRepo{
		db: db,
	}
}

func (r *ProjectRepo) InsertNewProject(ctx context.Context, tx pgx.Tx, modelProject *entity.ProjectEntity) (*entity.ProjectEntity, *app_errors.AppError) {
	cols := []string{"id", "name", "type", "visibility", "master_id"}
	vals := []any{modelProject.ID, modelProject.Name, modelProject.Type, modelProject.Visibility, modelProject.MasterID}

	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	projectQuery := fmt.Sprintf(`
	INSERT INTO projects (%s)
	VALUES (%s)
	RETURNING id, name, type, visibility, master_id, created_at;
	`, strings.Join(cols, ","), strings.Join(placeholders, ","))

	var project entity.ProjectEntity
	if err := tx.QueryRow(ctx, projectQuery, vals...).Scan(&project.ID, &project.Name, &project.Type, &project.Visibility, &project.MasterID, &project.CreatedAt); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	return &project, nil
}

func (r *ProjectRepo) InsertNewProjectMember(ctx context.Context, tx pgx.Tx, projectID, userID string, role entity.UserRole) *app_errors.AppError {

	projectMemberQuery := `
	INSERT INTO project_members (
		id, project_id, user_id, role
	) VALUES ($1, $2, $3, $4);
	`
	memberID, uuidErr := uuid.NewV7()
	if uuidErr != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", uuidErr)
	}

	if _, err := tx.Exec(ctx, projectMemberQuery, memberID.String(), projectID, userID, role); err != nil {
		return app_errors.MapPgxError(err)
	}
	return nil
}

func (r *ProjectRepo) GetSelfProject(ctx context.Context, userID string) ([]entity.ProjectSelf, *app_errors.AppError) {
	baseQuery := `
	SELECT p.id, p.name, p.type, p.visibility, p.master_id, pm.role 
	FROM project_members pm 
	JOIN projects p ON p.id = pm.project_id 
	WHERE pm.user_id = $1 
	ORDER BY p.created_at DESC;
	`

	var projects []entity.ProjectSelf
	rows, err := r.db.Query(ctx, baseQuery, userID)
	if err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p entity.ProjectSelf
		if err := rows.Scan(&p.ID, &p.Name, &p.Type, &p.Visibility, &p.MasterID, &p.Role); err != nil {
			return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return projects, nil
}

func (r *ProjectRepo) IsProjectExist(ctx context.Context, projectID string) (bool, *app_errors.AppError) {
	baseQuery := `
	SELECT EXISTS (
		SELECT 1
		FROM projects
		WHERE id = $1
	);
	`
	var exists bool
	if err := r.db.QueryRow(ctx, baseQuery, projectID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
		}
		return false, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return exists, nil
}

func (r *ProjectRepo) GetUserRoleInProject(ctx context.Context, userID, projectID string) (string, *app_errors.AppError) {
	baseQuery := `
	SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2;
	`
	var role string
	if err := r.db.QueryRow(ctx, baseQuery, projectID, userID).Scan(&role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", app_errors.NewAppError(fiber.StatusForbidden, app_errors.ErrForbidden, "forbidden", nil)
		}
		return "", app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	return role, nil
}

func (r *ProjectRepo) GetProjectByID(ctx context.Context, projectID string) (*entity.ProjectEntity, *app_errors.AppError) {
	baseQuery := `
	SELECT id, name, type, visibility, master_id, created_at FROM projects WHERE id = $1;
	`

	var project entity.ProjectEntity
	if err := r.db.QueryRow(ctx, baseQuery, projectID).Scan(&project.ID, &project.Name, &project.Type, &project.Visibility, &project.MasterID, &project.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return &project, nil
}

func (r *ProjectRepo) GetProjectMember(ctx context.Context, projectID string) ([]entity.ProjectMember, *app_errors.AppError) {
	query := `
	SELECT u.id, u.username, pm.project_id, pm.role, pm.joined_at
	FROM users u
	LEFT JOIN project_members pm ON pm.user_id = u.id
	LEFT JOIN projects p ON p.id = pm.project_id
	WHERE pm.project_id = $1;
	`
	var projectMembers []entity.ProjectMember
	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pm entity.ProjectMember
		if err := rows.Scan(&pm.UserID, &pm.Username, &pm.ProjectID, &pm.Role, &pm.JoinedAt); err != nil {
			return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
		}
		projectMembers = append(projectMembers, pm)
	}

	if err := rows.Err(); err != nil {
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	return projectMembers, nil
}

func (r *ProjectRepo) GetProjectMemberUserIDs(
	ctx context.Context,
	projectID string,
) (map[string]bool, *app_errors.AppError) {

	query := `
	SELECT user_id
	FROM project_members
	WHERE project_id = $1;
	`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer rows.Close()

	result := make(map[string]bool)

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, app_errors.MapPgxError(err)
		}
		result[userID] = true
	}

	return result, nil
}

func (r *ProjectRepo) GetPendingInvitations(ctx context.Context, projectID string) (map[string]bool, *app_errors.AppError) {
	// should get all user with pending invitations
	query := `
	SELECT invited_user_id 
	FROM project_invitations 
	WHERE project_id = $1
	AND status = 'Pending'
	AND expires_at > now();
	`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer rows.Close()

	result := make(map[string]bool)

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, app_errors.MapPgxError(err)
		}
		result[userID] = true
	}

	return result, nil
}

func (r *ProjectRepo) GetUsersByIds(ctx context.Context, userIDs []string) (map[string]bool, *app_errors.AppError) {

	if len(userIDs) == 0 {
		return map[string]bool{}, nil
	}

	query := `
	SELECT id
	FROM users
	WHERE id = ANY($1);
	`

	rows, err := r.db.Query(ctx, query, userIDs)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "user_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer rows.Close()

	result := make(map[string]bool)

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, app_errors.MapPgxError(err)
		}
		result[id] = true
	}

	return result, nil
}

func (r *ProjectRepo) BatchInsertProjectInvitation(ctx context.Context, tx pgx.Tx, invs []entity.ProjectInvitationEntity) *app_errors.AppError {
	query := `
	INSERT INTO project_invitations (
		id,
		project_id,
		invited_user_id,
		invited_by,
		role,
		status,
		token_hash,
		expires_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8);
	`

	batch := &pgx.Batch{}

	for _, inv := range invs {
		batch.Queue(query, inv.ID, inv.ProjectID, inv.InvitedUserID, inv.InvitedBy, inv.Role, inv.Status, inv.TokenHash, inv.ExpiresAt)
	}

	br := tx.SendBatch(ctx, batch)

	err := br.Close()
	if err != nil {
		return app_errors.MapPgxError(err)
	}

	return nil
}

func (r *ProjectRepo) GetInvitationInfo(ctx context.Context, invitationID string) (*entity.InvitationInfo, *app_errors.AppError) {
	query := `
	SELECT i.id, i.status, u.email, u.username, p.name FROM project_invitations i
	JOIN users u ON u.id = i.invited_user_id
	JOIN projects p ON p.id = i.project_id
	WHERE i.id = $1;
	`

	var resp entity.InvitationInfo
	if err := r.db.QueryRow(ctx, query, invitationID).Scan(&resp.ID, &resp.InvitationStatus, &resp.UserEmail, &resp.Username, &resp.ProjectName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return &resp, nil
}

func (r *ProjectRepo) GetInvitationProjectByID(ctx context.Context, invitationID string) (*entity.ProjectInvitationEntity, *app_errors.AppError) {
	query := `
	SELECT * FROM project_invitations WHERE id = $1;
	`

	var projectInvitation entity.ProjectInvitationEntity
	if err := r.db.QueryRow(ctx, query, invitationID).Scan(&projectInvitation.ID, &projectInvitation.ProjectID, &projectInvitation.InvitedUserID, &projectInvitation.InvitedBy, &projectInvitation.Role, &projectInvitation.Status, &projectInvitation.TokenHash, &projectInvitation.ExpiresAt, &projectInvitation.AcceptedAt, &projectInvitation.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return &projectInvitation, nil
}

func (r *ProjectRepo) UpdateUserInvitationState(ctx context.Context, tx pgx.Tx, invitationID, status string) *app_errors.AppError {
	query := `
		UPDATE project_invitations 
		SET accepted_at = now(), status = $1, token_hash = NULL
		WHERE id = $2;
	`
	if _, err := tx.Exec(ctx, query, status, invitationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "project_not_found", nil)
		}
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	return nil
}

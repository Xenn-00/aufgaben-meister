package user_repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Xenn-00/aufgaben-meister/internal/abstraction/tx"
	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) UserRepoContract {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) FindByUserID(ctx context.Context, userID string) (*entity.UserEntity, *app_errors.AppError) {
	// Base query
	query := `
		SELECT id, email, username, name, password_hash, created_at, updated_at FROM users WHERE id = $1 LIMIT 1
	`
	row := r.db.QueryRow(ctx, query, userID)

	var u entity.UserEntity
	if err := row.Scan(&u.ID, &u.Email, &u.Username, &u.Name, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "20000" {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "user_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	return &u, nil
}

func (r *UserRepo) IsUnderOneProject(ctx context.Context, reqUserID, userID string) (bool, *app_errors.AppError) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM project_members pm1
			JOIN project_members pm2
			ON pm1.project_id = pm2.project_id
			WHERE pm1.user_id = $1
			AND pm2.user_id = $2
		)
	`
	var exists bool
	err := r.db.QueryRow(ctx, query, reqUserID, userID).Scan(&exists)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "20000" {
			return false, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "user_not_found", nil)
		}
		return false, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	return exists, nil
}

func (r *UserRepo) FindUserWithProjects(ctx context.Context, userID string) (*entity.UserWithProject, *app_errors.AppError) {
	query := `
		SELECT
			u.id, u.email, u.username, u.name,
			p.id, p.name, p.type, p.visibility, pm.joined_at
		FROM users u
		LEFT JOIN project_members pm ON pm.user_id = u.id
		LEFT JOIN projects p ON p.id = pm.project_id
		WHERE u.id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "20000" {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "user_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}
	defer rows.Close()

	var userWithProject entity.UserWithProject
	projects := []entity.UserProject{}

	for rows.Next() {
		var p entity.UserProject
		rows.Scan(&userWithProject.ID, &userWithProject.Email, &userWithProject.Username, &userWithProject.Name, &p.ProjectID, &p.ProjectName, &p.ProjectType, &p.ProjectVisibility, &p.JoinedAtProject)
		projects = append(projects, p)
	}

	userWithProject.Projects = projects
	return &userWithProject, nil
}

func (r *UserRepo) UpdateSelfProfileTx(ctx context.Context, t tx.Tx, userID string, model entity.UserUpdate) (*entity.UserEntity, *app_errors.AppError) {
	pgxTx := t.(*tx.PgxTx).Tx
	setClauses := make([]string, 0)
	args := make([]any, 0)
	argPos := 1

	if model.Username != "" {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argPos))
		args = append(args, model.Username)
		argPos++
	}

	if model.Name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argPos))
		args = append(args, model.Name)
		argPos++
	}

	if model.Email != "" {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argPos))
		args = append(args, model.Email)
		argPos++
	}

	if model.UpdatedAt != "" {
		setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argPos))
		args = append(args, model.UpdatedAt)
		argPos++
	}

	if len(setClauses) == 0 {
		return nil, app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", fmt.Errorf("Keine Felder zum Aktualisieren."))
	}

	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $%d
		RETURNING id, username, name, email, created_at, updated_at
	`, strings.Join(setClauses, ", "), argPos)

	args = append(args, userID)

	row := pgxTx.QueryRow(ctx, query, args...)

	var user entity.UserEntity
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "user_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return &user, nil
}

func (r *UserRepo) DeactivateSelfUser(ctx context.Context, t tx.Tx, userID string) (bool, *app_errors.AppError) {
	pgxTx := t.(*tx.PgxTx).Tx
	query := `
		UPDATE users
		SET is_active = false
		WHERE id = $1
	`
	if _, err := pgxTx.Exec(ctx, query, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "user_not_found", nil)
		}
		return false, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return true, nil
}

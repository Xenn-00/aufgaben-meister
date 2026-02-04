package auth_repo

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

type AuthRepo struct {
	db *pgxpool.Pool
}

// Hier können die Methoden des AuthRepo implementiert werden.
func NewAuthRepo(db *pgxpool.Pool) AuthRepoContract {
	return &AuthRepo{
		db: db,
	}
}

func (r *AuthRepo) CountUsers(ctx context.Context, filter entity.UserCountFilter) (int64, *app_errors.AppError) {
	var count int64

	// Base query
	query := `SELECT COUNT(*) FROM users WHERE 1=1` // 1=1 dient als Platzhalter für einfache Erweiterungen
	args := []any{}                                 // Argumente für die Abfrage
	argPos := 1                                     // Position der Argumente für pgx ($1, $2, ...)

	// Dynamische Filter hinzufügen
	if filter.Email != nil {
		query += fmt.Sprintf(" AND email = $%d", argPos)
		args = append(args, *filter.Email)
		argPos++
	}

	if filter.Username != nil {
		query += fmt.Sprintf(" AND username = $%d", argPos)
		args = append(args, *filter.Username)
		argPos++
	}

	// Abfrage ausführen
	err := r.db.QueryRow(ctx, query, args...).Scan(&count) // Scan das Ergebnis in die count Variable
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil // Kein Benutzer gefunden
		}
		return 0, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return count, nil
}

// SaveUsers speichert einen neuen Benutzer in der Datenbank.
// Nimmt einen Kontext und ein entity.UserEntity entgegen und gibt die erzeugte Benutzer-ID,
// die zugewiesene Rolle sowie einen optionalen AppError zurück.
func (r *AuthRepo) SaveUsers(ctx context.Context, model entity.UserEntity) (string, *app_errors.AppError) {

	cols := []string{"id", "email", "password_hash", "name", "username"}
	vals := []any{model.ID, model.Email, model.PasswordHash, model.Name, model.Username}

	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	// Base query
	query := fmt.Sprintf(`
	INSERT INTO users (%s)
	VALUES (%s)
	RETURNING id;
	`, strings.Join(cols, ","), strings.Join(placeholders, ","))

	var id string
	if err := r.db.QueryRow(ctx, query, vals...).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		// Bei Verletzung von Unique-Constraints (z.B. E-Mail oder Username) wird ein Konfliktfehler (StatusConflict) zurückgegeben.
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", app_errors.NewAppError(fiber.StatusConflict, app_errors.ErrConflict, "conflict", nil)
		}
		return "", app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return id, nil
}

// FindByEmail sucht einen Benutzer anhand der übergebenen E‑Mail-Adresse.
// Der Kontext ctx steuert die Anfrage, email ist die gesuchte E‑Mail.
// Liefert bei Erfolg die *entity.UserEntity zurück. Bei Nichtexistenz oder DB-Fehlern
// wird ein *app_errors.AppError mit passendem Status (z. B. NotFound oder InternalServerError) zurückgegeben.
func (r *AuthRepo) FindByEmail(ctx context.Context, email string) (*entity.UserEntity, *app_errors.AppError) {
	// Base query
	query := `
		SELECT id, email, username, password_hash, is_active FROM users WHERE email = $1 LIMIT 1
	`
	row := r.db.QueryRow(ctx, query, email)

	var u entity.UserEntity
	if err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.IsActive); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "20000" {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "user_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return &u, nil
}

// FindByUsername sucht einen Benutzer anhand der übergebenen Benutzername
// Der Kontext ctx steuert die Anfrage, username ist der gesuchte Benutzername.
// Liefert bei Erfolg die *entity.UserEntity zurück. Bei Nichtexistenz oder DB-Fehlern
// wird ein *app_errors.AppError mit passendem Status (z. B. NotFound oder InternalServerError) zurückgegeben.
func (r *AuthRepo) FindByUsername(ctx context.Context, username string) (*entity.UserEntity, *app_errors.AppError) {
	// Base query
	query := `
		SELECT id, email, username, password_hash, is_active FROM users WHERE username = $1 LIMIT 1
	`
	row := r.db.QueryRow(ctx, query, username)

	var u entity.UserEntity
	if err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.IsActive); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "20000" {
			return nil, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "user_not_found", nil)
		}
		return nil, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return &u, nil
}

func (r *AuthRepo) IsUserActive(ctx context.Context, userID string) (bool, *app_errors.AppError) {
	query := `
		SELECT is_active FROM users WHERE id = $1 
	`
	var IsActive bool
	err := r.db.QueryRow(ctx, query, userID).Scan(&IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, app_errors.NewAppError(fiber.StatusNotFound, app_errors.ErrNotFound, "user_not_found", nil)
		}
		return false, app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "internal_error", err)
	}

	return IsActive, nil
}

func (r *AuthRepo) UserActivate(ctx context.Context, t tx.Tx, userID string) (bool, *app_errors.AppError) {
	pgxTx := t.(*tx.PgxTx).Tx
	query := `
		UPDATE users
		SET is_active = true
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

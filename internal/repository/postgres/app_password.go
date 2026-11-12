package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/larssonoliver/inundated/internal/model"
)

// ListAppPasswords implements [repository.AppPasswordRepository].
func (r *PostgresStore) ListAppPasswords(ctx context.Context, scope model.OwnerScope) ([]model.AppPassword, error) {
	ownerSQL, args := ownerPredicate("user_id", scope, []any{})
	q := `
		SELECT id, user_id, name, created_at, last_used_at
		FROM app_passwords
		WHERE ` + ownerSQL + `
		ORDER BY created_at DESC, id`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("ListAppPasswords: %w", err)
	}
	defer rows.Close()

	passwords := []model.AppPassword{}
	for rows.Next() {
		var p model.AppPassword
		if err := rows.Scan(&p.Id, &p.UserId, &p.Name, &p.CreatedAt, &p.LastUsedAt); err != nil {
			return nil, fmt.Errorf("ListAppPasswords: %w", err)
		}
		passwords = append(passwords, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListAppPasswords: %w", err)
	}
	return passwords, nil
}

// CreateAppPassword implements [repository.AppPasswordRepository].
func (r *PostgresStore) CreateAppPassword(ctx context.Context, scope model.OwnerScope, password model.AppPassword, token string) (model.AppPassword, error) {
	if token == "" {
		return model.AppPassword{}, fmt.Errorf("CreateAppPassword: token must not be empty: %w", model.ErrInvalidArgument)
	}
	if password.Id == uuid.Nil {
		password.Id = uuid.New()
	}
	if password.CreatedAt.IsZero() {
		password.CreatedAt = time.Now()
	}

	const q = `
		INSERT INTO app_passwords (id, user_id, name, token_hash, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, name, created_at, last_used_at`

	var created model.AppPassword
	err := r.db.QueryRow(ctx, q, password.Id, scope.UserID(), password.Name, model.HashToken(token), password.CreatedAt).
		Scan(&created.Id, &created.UserId, &created.Name, &created.CreatedAt, &created.LastUsedAt)
	if isUniqueViolation(err) {
		return model.AppPassword{}, fmt.Errorf("CreateAppPassword %s: %w", password.Id, model.ErrAlreadyExists)
	}
	if err != nil {
		return model.AppPassword{}, fmt.Errorf("CreateAppPassword: %w", err)
	}
	return created, nil
}

// DeleteAppPassword implements [repository.AppPasswordRepository].
func (r *PostgresStore) DeleteAppPassword(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	ownerSQL, args := ownerPredicate("user_id", scope, []any{id})
	q := `DELETE FROM app_passwords WHERE id = $1 AND ` + ownerSQL

	res, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("DeleteAppPassword: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("DeleteAppPassword %s: %w", id, model.ErrNotFound)
	}
	return nil
}

// GetAppPasswordByToken implements [repository.AppPasswordRepository].
func (r *PostgresStore) GetAppPasswordByToken(ctx context.Context, token string) (model.AppPassword, error) {
	const q = `
		SELECT id, user_id, name, created_at, last_used_at
		FROM app_passwords
		WHERE token_hash = $1`

	var p model.AppPassword
	err := r.db.QueryRow(ctx, q, model.HashToken(token)).
		Scan(&p.Id, &p.UserId, &p.Name, &p.CreatedAt, &p.LastUsedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.AppPassword{}, fmt.Errorf("GetAppPasswordByToken: %w", model.ErrNotFound)
	}
	if err != nil {
		return model.AppPassword{}, fmt.Errorf("GetAppPasswordByToken: %w", err)
	}
	return p, nil
}

// TouchAppPassword implements [repository.AppPasswordRepository].
func (r *PostgresStore) TouchAppPassword(ctx context.Context, id uuid.UUID, usedAt time.Time) error {
	const q = `UPDATE app_passwords SET last_used_at = $2 WHERE id = $1`

	res, err := r.db.Exec(ctx, q, id, usedAt)
	if err != nil {
		return fmt.Errorf("TouchAppPassword: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("TouchAppPassword %s: %w", id, model.ErrNotFound)
	}
	return nil
}

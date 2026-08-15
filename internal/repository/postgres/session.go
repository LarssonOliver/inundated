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

// CreateSession implements [repository.SessionRepository].
func (r *PostgresStore) CreateSession(ctx context.Context, session model.Session) (model.Session, error) {
	if session.UserId == uuid.Nil {
		return model.Session{}, fmt.Errorf("CreateSession: user_id: %w", model.ErrInvalidArgument)
	}
	if session.Sub == "" {
		return model.Session{}, fmt.Errorf("CreateSession: sub must not be empty: %w", model.ErrInvalidArgument)
	}
	if session.Id == uuid.Nil {
		session.Id = uuid.New()
	}

	const q = `
		INSERT INTO sessions (id, user_id, sub, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, sub, expires_at`

	var created model.Session
	err := r.db.QueryRow(ctx, q, session.Id, session.UserId, session.Sub, session.ExpiresAt).
		Scan(&created.Id, &created.UserId, &created.Sub, &created.ExpiresAt)
	if isUniqueViolation(err) {
		return model.Session{}, fmt.Errorf("CreateSession %s: %w", session.Id, model.ErrAlreadyExists)
	}
	if err != nil {
		return model.Session{}, fmt.Errorf("CreateSession: %w", err)
	}
	return created, nil
}

// DeleteSession implements [repository.SessionRepository].
func (r *PostgresStore) DeleteSession(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("DeleteSession: id: %w", model.ErrInvalidArgument)
	}

	const q = `DELETE FROM sessions WHERE id = $1`

	res, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("DeleteSession: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("DeleteSession %s: %w", id, model.ErrNotFound)
	}
	return nil
}

// GetSession implements [repository.SessionRepository].
func (r *PostgresStore) GetSession(ctx context.Context, id uuid.UUID) (model.Session, error) {
	if id == uuid.Nil {
		return model.Session{}, fmt.Errorf("GetSession: id: %w", model.ErrInvalidArgument)
	}

	const q = `
		SELECT id, user_id, sub, expires_at
		FROM sessions
		WHERE id = $1`

	var s model.Session
	err := r.db.QueryRow(ctx, q, id).Scan(&s.Id, &s.UserId, &s.Sub, &s.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Session{}, fmt.Errorf("GetSession %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.Session{}, fmt.Errorf("GetSession: %w", err)
	}
	return s, nil
}

// TouchSession implements [repository.SessionRepository].
func (r *PostgresStore) TouchSession(ctx context.Context, id uuid.UUID, expiresAt time.Time) (model.Session, error) {
	if id == uuid.Nil {
		return model.Session{}, fmt.Errorf("TouchSession: id: %w", model.ErrInvalidArgument)
	}

	const q = `
		UPDATE sessions
		SET expires_at = $2
		WHERE id = $1
		RETURNING id, user_id, sub, expires_at`

	var updated model.Session
	err := r.db.QueryRow(ctx, q, id, expiresAt).
		Scan(&updated.Id, &updated.UserId, &updated.Sub, &updated.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Session{}, fmt.Errorf("TouchSession %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.Session{}, fmt.Errorf("TouchSession: %w", err)
	}
	return updated, nil
}

// DeleteAllExpiredSessions implements [repository.SessionRepository].
func (r *PostgresStore) DeleteAllExpiredSessions(ctx context.Context) error {
	const q = "DELETE FROM sessions WHERE expires_at < NOW()"

	_, err := r.db.Exec(ctx, q)

	if err != nil {
		return fmt.Errorf("DeleteAllExpiredSessions: %w", err)
	}

	return nil
}

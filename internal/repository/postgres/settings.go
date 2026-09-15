package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/larssonoliver/inundated/internal/model"
)

// GetSettings implements [repository.SettingsRepository].
func (r *PostgresStore) GetSettings(ctx context.Context, scope model.OwnerScope) (model.Settings, error) {
	ownerSQL, args := ownerPredicate("user_id", scope, []any{})
	q := `
		SELECT id, user_id, week_start_day, timezone, duration_format, time_format
		FROM settings
		WHERE ` + ownerSQL

	var s model.Settings
	err := r.db.QueryRow(ctx, q, args...).
		Scan(&s.Id, &s.UserId, &s.WeekStartDay, &s.Timezone, &s.DurationFormat, &s.TimeFormat)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Settings{}, fmt.Errorf("GetSettings: %w", model.ErrNotFound)
	}
	if err != nil {
		return model.Settings{}, fmt.Errorf("GetSettings: %w", err)
	}
	return s, nil
}

// CreateSettings implements [repository.SettingsRepository].
func (r *PostgresStore) CreateSettings(ctx context.Context, scope model.OwnerScope, settings model.Settings) (model.Settings, error) {
	if settings.Id == uuid.Nil {
		settings.Id = uuid.New()
	}

	const q = `
		INSERT INTO settings (id, user_id, week_start_day, timezone, duration_format, time_format)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, week_start_day, timezone, duration_format, time_format`

	var created model.Settings
	err := r.db.QueryRow(ctx, q,
		settings.Id, scope.UserID(), settings.WeekStartDay, settings.Timezone, settings.DurationFormat, settings.TimeFormat,
	).Scan(&created.Id, &created.UserId, &created.WeekStartDay, &created.Timezone, &created.DurationFormat, &created.TimeFormat)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Settings{}, fmt.Errorf("CreateSettings: %w", model.ErrAlreadyExists)
		}
		return model.Settings{}, fmt.Errorf("CreateSettings: %w", err)
	}
	return created, nil
}

// UpdateSettings implements [repository.SettingsRepository].
func (r *PostgresStore) UpdateSettings(ctx context.Context, scope model.OwnerScope, settings model.Settings) (model.Settings, error) {
	ownerSQL, args := ownerPredicate("user_id", scope, []any{
		settings.WeekStartDay, settings.Timezone, settings.DurationFormat, settings.TimeFormat,
	})
	q := `
		UPDATE settings
		SET week_start_day = $1, timezone = $2, duration_format = $3, time_format = $4
		WHERE ` + ownerSQL + `
		RETURNING id, user_id, week_start_day, timezone, duration_format, time_format`

	var updated model.Settings
	err := r.db.QueryRow(ctx, q, args...).
		Scan(&updated.Id, &updated.UserId, &updated.WeekStartDay, &updated.Timezone, &updated.DurationFormat, &updated.TimeFormat)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Settings{}, fmt.Errorf("UpdateSettings: %w", model.ErrNotFound)
	}
	if err != nil {
		return model.Settings{}, fmt.Errorf("UpdateSettings: %w", err)
	}
	return updated, nil
}

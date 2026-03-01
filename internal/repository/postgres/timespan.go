package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/larssonoliver/inundated/internal/model"
)

func (r *PostgresStore) GetTimeSpan(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
	if id == uuid.Nil {
		return model.TimeSpan{}, fmt.Errorf("GetTimeSpan: id: %w", model.ErrInvalidArgument)
	}

	const q = `SELECT id, name, start_time, end_time FROM timespans WHERE id = $1`

	var ts model.TimeSpan
	err := r.db.QueryRow(ctx, q, id).Scan(&ts.Id, &ts.Name, &ts.StartTime, &ts.EndTime)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.TimeSpan{}, fmt.Errorf("GetTimeSpan %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.TimeSpan{}, fmt.Errorf("GetTimeSpan: %w", err)
	}

	ts.TagIds, err = r.timeSpanTagIds(ctx, id)
	if err != nil {
		return model.TimeSpan{}, err
	}
	return ts, nil
}

func (r *PostgresStore) ListTimeSpans(ctx context.Context) ([]model.TimeSpan, error) {
	const q = `SELECT id, name, start_time, end_time FROM timespans ORDER BY start_time DESC`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("ListTimeSpans: %w", err)
	}
	defer rows.Close()

	var spans []model.TimeSpan
	for rows.Next() {
		var ts model.TimeSpan
		if err := rows.Scan(&ts.Id, &ts.Name, &ts.StartTime, &ts.EndTime); err != nil {
			return nil, fmt.Errorf("ListTimeSpans scan: %w", err)
		}
		spans = append(spans, ts)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListTimeSpans rows: %w", err)
	}

	for i := range spans {
		spans[i].TagIds, err = r.timeSpanTagIds(ctx, spans[i].Id)
		if err != nil {
			return nil, err
		}
	}
	return spans, nil
}

func (r *PostgresStore) CreateTimeSpan(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
	if timeSpan.StartTime.IsZero() {
		return model.TimeSpan{}, fmt.Errorf("CreateTimeSpan: start_time must not be zero: %w", model.ErrInvalidArgument)
	}
	if !timeSpan.EndTime.IsZero() && !timeSpan.EndTime.After(timeSpan.StartTime) {
		return model.TimeSpan{}, fmt.Errorf("CreateTimeSpan: end_time must be after start_time: %w", model.ErrInvalidArgument)
	}
	if timeSpan.Id == uuid.Nil {
		timeSpan.Id = uuid.New()
	}

	const q = `
		INSERT INTO timespans (id, name, start_time, end_time)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, start_time, end_time`

	var created model.TimeSpan
	err := r.db.QueryRow(ctx, q, timeSpan.Id, timeSpan.Name, timeSpan.StartTime, timeSpan.EndTime).
		Scan(&created.Id, &created.Name, &created.StartTime, &created.EndTime)
	if err != nil {
		return model.TimeSpan{}, fmt.Errorf("CreateTimeSpan: %w", err)
	}

	if err := r.setTimeSpanTags(ctx, created.Id, timeSpan.TagIds); err != nil {
		return model.TimeSpan{}, err
	}
	created.TagIds = timeSpan.TagIds
	return created, nil
}

func (r *PostgresStore) UpdateTimeSpan(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
	if timeSpan.Id == uuid.Nil {
		return model.TimeSpan{}, fmt.Errorf("UpdateTimeSpan: id: %w", model.ErrInvalidArgument)
	}
	if timeSpan.StartTime.IsZero() {
		return model.TimeSpan{}, fmt.Errorf("UpdateTimeSpan: start_time must not be zero: %w", model.ErrInvalidArgument)
	}
	if !timeSpan.EndTime.IsZero() && !timeSpan.EndTime.After(timeSpan.StartTime) {
		return model.TimeSpan{}, fmt.Errorf("UpdateTimeSpan: end_time must be after start_time: %w", model.ErrInvalidArgument)
	}

	const q = `
		UPDATE timespans SET name = $2, start_time = $3, end_time = $4
		WHERE id = $1
		RETURNING id, name, start_time, end_time`

	var updated model.TimeSpan
	err := r.db.QueryRow(ctx, q, timeSpan.Id, timeSpan.Name, timeSpan.StartTime, timeSpan.EndTime).
		Scan(&updated.Id, &updated.Name, &updated.StartTime, &updated.EndTime)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.TimeSpan{}, fmt.Errorf("UpdateTimeSpan %s: %w", timeSpan.Id, model.ErrNotFound)
	}
	if err != nil {
		return model.TimeSpan{}, fmt.Errorf("UpdateTimeSpan: %w", err)
	}

	if err := r.setTimeSpanTags(ctx, updated.Id, timeSpan.TagIds); err != nil {
		return model.TimeSpan{}, err
	}
	updated.TagIds = timeSpan.TagIds
	return updated, nil
}

func (r *PostgresStore) DeleteTimeSpan(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("DeleteTimeSpan: id: %w", model.ErrInvalidArgument)
	}

	const q = `DELETE FROM timespans WHERE id = $1`

	res, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("DeleteTimeSpan: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("DeleteTimeSpan %s: %w", id, model.ErrNotFound)
	}
	return nil
}

// timeSpanTagIds returns all tag IDs linked to a time span.
func (r *PostgresStore) timeSpanTagIds(ctx context.Context, timeSpanId uuid.UUID) ([]uuid.UUID, error) {
	const q = `SELECT tag_id FROM timespan_tags WHERE timespan_id = $1 ORDER BY tag_id`

	rows, err := r.db.Query(ctx, q, timeSpanId)
	if err != nil {
		return nil, fmt.Errorf("timeSpanTagIds: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("timeSpanTagIds scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// setTimeSpanTags replaces all tag associations for a time span.
func (r *PostgresStore) setTimeSpanTags(ctx context.Context, timeSpanId uuid.UUID, tagIds []uuid.UUID) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM timespan_tags WHERE timespan_id = $1`, timeSpanId); err != nil {
		return fmt.Errorf("setTimeSpanTags delete: %w", err)
	}
	for _, tagId := range tagIds {
		if _, err := r.db.Exec(ctx,
			`INSERT INTO timespan_tags (timespan_id, tag_id) VALUES ($1, $2)`,
			timeSpanId, tagId,
		); err != nil {
			return fmt.Errorf("setTimeSpanTags insert: %w", model.ErrInvalidReference)
		}
	}
	return nil
}

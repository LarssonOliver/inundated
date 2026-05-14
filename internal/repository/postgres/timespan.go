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

func (r *PostgresStore) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	if id == uuid.Nil {
		return model.Timespan{}, fmt.Errorf("GetTimespan: id: %w", model.ErrInvalidArgument)
	}

	const q = `
		SELECT id, name, start_time, end_time 
		FROM timespans 
		WHERE id = $1 AND deleted_at IS NULL`

	var ts model.Timespan
	err := r.db.QueryRow(ctx, q, id).Scan(&ts.Id, &ts.Name, &ts.StartTime, &ts.EndTime)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Timespan{}, fmt.Errorf("GetTimespan %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.Timespan{}, fmt.Errorf("GetTimespan: %w", err)
	}

	ts.TagIds, err = r.timespanTagIds(ctx, id)
	if err != nil {
		return model.Timespan{}, err
	}
	return ts, nil
}

func (r *PostgresStore) ListTimespans(ctx context.Context) ([]model.Timespan, error) {
	const q = `
		SELECT id, name, start_time, end_time 
		FROM timespans 
		WHERE deleted_at IS NULL
		ORDER BY start_time DESC`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("ListTimespans: %w", err)
	}
	defer rows.Close()

	var spans []model.Timespan
	for rows.Next() {
		var ts model.Timespan
		if err := rows.Scan(&ts.Id, &ts.Name, &ts.StartTime, &ts.EndTime); err != nil {
			return nil, fmt.Errorf("ListTimespans scan: %w", err)
		}
		spans = append(spans, ts)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListTimespans rows: %w", err)
	}

	for i := range spans {
		spans[i].TagIds, err = r.timespanTagIds(ctx, spans[i].Id)
		if err != nil {
			return nil, err
		}
	}
	return spans, nil
}

func (r *PostgresStore) CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	if timespan.StartTime.IsZero() {
		return model.Timespan{}, fmt.Errorf("CreateTimespan: start_time must not be zero: %w", model.ErrInvalidArgument)
	}
	if !timespan.EndTime.IsZero() && !timespan.EndTime.After(timespan.StartTime) {
		return model.Timespan{}, fmt.Errorf("CreateTimespan: end_time must be after start_time: %w", model.ErrInvalidArgument)
	}
	if timespan.Id == uuid.Nil {
		timespan.Id = uuid.New()
	}

	const q = `
		INSERT INTO timespans (id, name, start_time, end_time)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, start_time, end_time`

	var created model.Timespan
	err := r.db.QueryRow(ctx, q, timespan.Id, timespan.Name, timespan.StartTime, timespan.EndTime).
		Scan(&created.Id, &created.Name, &created.StartTime, &created.EndTime)
	if err != nil {
		return model.Timespan{}, fmt.Errorf("CreateTimespan: %w", err)
	}

	if err := r.setTimespanTags(ctx, created.Id, timespan.TagIds); err != nil {
		return model.Timespan{}, err
	}
	created.TagIds = timespan.TagIds
	return created, nil
}

func (r *PostgresStore) UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	if timespan.Id == uuid.Nil {
		return model.Timespan{}, fmt.Errorf("UpdateTimespan: id: %w", model.ErrInvalidArgument)
	}
	if timespan.StartTime.IsZero() {
		return model.Timespan{}, fmt.Errorf("UpdateTimespan: start_time must not be zero: %w", model.ErrInvalidArgument)
	}
	if !timespan.EndTime.IsZero() && !timespan.EndTime.After(timespan.StartTime) {
		return model.Timespan{}, fmt.Errorf("UpdateTimespan: end_time must be after start_time: %w", model.ErrInvalidArgument)
	}

	const q = `
		UPDATE timespans SET name = $2, start_time = $3, end_time = $4
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, start_time, end_time`

	var updated model.Timespan
	err := r.db.QueryRow(ctx, q, timespan.Id, timespan.Name, timespan.StartTime, timespan.EndTime).
		Scan(&updated.Id, &updated.Name, &updated.StartTime, &updated.EndTime)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Timespan{}, fmt.Errorf("UpdateTimespan %s: %w", timespan.Id, model.ErrNotFound)
	}
	if err != nil {
		return model.Timespan{}, fmt.Errorf("UpdateTimespan: %w", err)
	}

	if err := r.setTimespanTags(ctx, updated.Id, timespan.TagIds); err != nil {
		return model.Timespan{}, err
	}
	updated.TagIds = timespan.TagIds
	return updated, nil
}

func (r *PostgresStore) DeleteTimespan(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("DeleteTimespan: id: %w", model.ErrInvalidArgument)
	}

	const q = `
		UPDATE timespans 
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL`

	res, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("DeleteTimespan: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("DeleteTimespan %s: %w", id, model.ErrNotFound)
	}
	return nil
}

// timespanTagIds returns all tag IDs linked to a time span.
func (r *PostgresStore) timespanTagIds(ctx context.Context, timespanId uuid.UUID) ([]uuid.UUID, error) {
	const q = `
		SELECT tag_id 
		FROM timespan_tags 
		WHERE timespan_id = $1 
		ORDER BY tag_id`

	rows, err := r.db.Query(ctx, q, timespanId)
	if err != nil {
		return nil, fmt.Errorf("timespanTagIds: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("timespanTagIds scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// setTimespanTags replaces all tag associations for a time span.
func (r *PostgresStore) setTimespanTags(ctx context.Context, timespanId uuid.UUID, tagIds []uuid.UUID) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM timespan_tags WHERE timespan_id = $1`, timespanId); err != nil {
		return fmt.Errorf("setTimespanTags delete: %w", err)
	}
	for _, tagId := range tagIds {
		if _, err := r.db.Exec(ctx,
			`INSERT INTO timespan_tags (timespan_id, tag_id) VALUES ($1, $2)`,
			timespanId, tagId,
		); err != nil {
			return fmt.Errorf("setTimespanTags insert: %w", model.ErrInvalidReference)
		}
	}
	return nil
}

// GetTotalDurationByTags implements [repository.Repository].
func (r *PostgresStore) GetTotalDurationByTags(ctx context.Context, tagIds []uuid.UUID) (time.Duration, error) {
	if len(tagIds) == 0 {
		return 0, nil
	}

	const q = `
		SELECT
			SUM(t.end_time - t.start_time) AS total_time
		FROM timespans t
		WHERE t.deleted_at IS NULL AND EXISTS (
			SELECT 1 FROM timespan_tags tt
			WHERE tt.timespan_id = t.id AND tt.tag_id = ANY($1)
		)`
	
	var duration *time.Duration
	err := r.db.QueryRow(ctx, q, tagIds).Scan(&duration)
	if errors.Is(err, pgx.ErrNoRows) || duration == nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("GetTotalDurationByTags: %w", err)
	}

	return *duration, nil
}

// AggregateTimeSpentByTagsAndBuckets implements [repository.ProjectStatsRepository].
func (r *PostgresStore) AggregateTimeSpentByTagsAndBuckets(ctx context.Context, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error) {
	values := make([]model.BucketValue, len(buckets))
	for i, bucket := range buckets {
		if !bucket.End.After(bucket.Start) {
			return nil, fmt.Errorf("AggregateTimeSpentByTagsAndBuckets: invalid bucket range: %w", model.ErrInvalidArgument)
		}

		values[i] = model.BucketValue{
			Bucket: bucket,
			Value:  0,
		}
	}

	if len(buckets) == 0 || len(tagIds) == 0 {
		return values, nil
	}

	bucketStarts := make([]time.Time, len(buckets))
	bucketEnds := make([]time.Time, len(buckets))
	for i, bucket := range buckets {
		bucketStarts[i] = bucket.Start
		bucketEnds[i] = bucket.End
	}

	const q = `
		WITH input_buckets AS (
			SELECT b_start, b_end, ord
			FROM unnest($2::timestamptz[], $3::timestamptz[]) WITH ORDINALITY AS b(b_start, b_end, ord)
		),
		matching_timespans AS (
			SELECT t.id, t.start_time, t.end_time
			FROM timespans t
			WHERE t.deleted_at IS NULL AND EXISTS (
				SELECT 1 FROM timespan_tags tt
				WHERE tt.timespan_id = t.id AND tt.tag_id = ANY($1)
			)
		)
		SELECT
			ib.b_start AS bucket_start,
			ib.b_end AS bucket_end,
			COALESCE(
				SUM(
					EXTRACT(EPOCH FROM (LEAST(mt.end_time, ib.b_end) - GREATEST(mt.start_time, ib.b_start)))
				) FILTER (WHERE mt.start_time < ib.b_end AND mt.end_time > ib.b_start),
				0
			) AS value_seconds
		FROM input_buckets ib
		LEFT JOIN matching_timespans mt
			ON mt.start_time < ib.b_end AND mt.end_time > ib.b_start
		GROUP BY ib.ord, ib.b_start, ib.b_end
		ORDER BY ib.ord`

	rows, err := r.db.Query(ctx, q, tagIds, bucketStarts, bucketEnds)
	if err != nil {
		return nil, fmt.Errorf("AggregateTimeSpentByTagsAndBuckets: %w", err)
	}
	defer rows.Close()

	result := make([]model.BucketValue, 0, len(buckets))
	for rows.Next() {
		var start time.Time
		var end time.Time
		var valueSeconds float64

		if err := rows.Scan(&start, &end, &valueSeconds); err != nil {
			return nil, fmt.Errorf("AggregateTimeSpentByTagsAndBuckets scan: %w", err)
		}

		result = append(result, model.BucketValue{
			Bucket: model.BucketRange{
				Start: start,
				End:   end,
			},
			Value: valueSeconds,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("AggregateTimeSpentByTagsAndBuckets rows: %w", err)
	}

	return result, nil
}

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/larssonoliver/inundated/internal/model"
)

func (r *PostgresStore) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	if id == uuid.Nil {
		return model.Tag{}, fmt.Errorf("GetTag: id: %w", model.ErrInvalidArgument)
	}

	const q = `SELECT id, name, color FROM tags WHERE id = $1`

	var t model.Tag
	err := r.db.QueryRow(ctx, q, id).Scan(&t.Id, &t.Name, &t.Color)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Tag{}, fmt.Errorf("GetTag %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.Tag{}, fmt.Errorf("GetTag: %w", err)
	}
	return t, nil
}

func (r *PostgresStore) ListTags(ctx context.Context) ([]model.Tag, error) {
	const q = `SELECT id, name, color FROM tags ORDER BY name`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("ListTags: %w", err)
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.Id, &t.Name, &t.Color); err != nil {
			return nil, fmt.Errorf("ListTags scan: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListTags rows: %w", err)
	}
	return tags, nil
}

func (r *PostgresStore) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	if tag.Name == "" {
		return model.Tag{}, fmt.Errorf("CreateTag: name must not be empty: %w", model.ErrInvalidArgument)
	}
	if tag.Id == uuid.Nil {
		tag.Id = uuid.New()
	}

	const q = `
		INSERT INTO tags (id, name, color)
		VALUES ($1, $2, $3)
		RETURNING id, name, color`

	var created model.Tag
	err := r.db.QueryRow(ctx, q, tag.Id, tag.Name, tag.Color).
		Scan(&created.Id, &created.Name, &created.Color)
	if err != nil {
		return model.Tag{}, fmt.Errorf("CreateTag: %w", err)
	}
	return created, nil
}

func (r *PostgresStore) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	if tag.Id == uuid.Nil {
		return model.Tag{}, fmt.Errorf("UpdateTag: id: %w", model.ErrInvalidArgument)
	}
	if tag.Name == "" {
		return model.Tag{}, fmt.Errorf("UpdateTag: name must not be empty: %w", model.ErrInvalidArgument)
	}

	const q = `
		UPDATE tags SET name = $2, color = $3
		WHERE id = $1
		RETURNING id, name, color`

	var updated model.Tag
	err := r.db.QueryRow(ctx, q, tag.Id, tag.Name, tag.Color).
		Scan(&updated.Id, &updated.Name, &updated.Color)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Tag{}, fmt.Errorf("UpdateTag %s: %w", tag.Id, model.ErrNotFound)
	}
	if err != nil {
		return model.Tag{}, fmt.Errorf("UpdateTag: %w", err)
	}
	return updated, nil
}

func (r *PostgresStore) DeleteTag(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("DeleteTag: id: %w", model.ErrInvalidArgument)
	}

	const q = `DELETE FROM tags WHERE id = $1`

	res, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("DeleteTag: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("DeleteTag %s: %w", id, model.ErrNotFound)
	}
	return nil
}

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/larssonoliver/inundated/internal/model"
)

func (r *PostgresStore) GetTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Tag, error) {
	if id == uuid.Nil {
		return model.Tag{}, fmt.Errorf("GetTag: id: %w", model.ErrInvalidArgument)
	}

	const q = `
		SELECT id, name, color, user_id
		FROM tags
		WHERE id = $1 AND deleted_at IS NULL AND user_id IS NOT DISTINCT FROM $2`

	var t model.Tag
	err := r.db.QueryRow(ctx, q, id, scope.UserID()).Scan(&t.Id, &t.Name, &t.Color, &t.UserId)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Tag{}, fmt.Errorf("GetTag %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.Tag{}, fmt.Errorf("GetTag: %w", err)
	}
	return t, nil
}

func (r *PostgresStore) ListTags(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Tag], error) {
	ownerSQL, ownerArgs := ownerPredicate("user_id", scope, 1)
	countQ := `
		SELECT COUNT(*)
		FROM tags
		WHERE deleted_at IS NULL AND ` + ownerSQL

	var totalCount int
	if err := r.db.QueryRow(ctx, countQ, ownerArgs...).Scan(&totalCount); err != nil {
		return model.Page[model.Tag]{}, fmt.Errorf("count tags: %w", err)
	}

	dataOwnerSQL, _ := ownerPredicate("user_id", scope, 3)
	q := `
		SELECT id, name, color, user_id
		FROM tags
		WHERE deleted_at IS NULL AND ` + dataOwnerSQL + `
		ORDER BY name
		LIMIT $1 OFFSET $2`

	args := append([]any{params.Limit, params.Offset}, ownerArgs...)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return model.Page[model.Tag]{}, fmt.Errorf("ListTags: %w", err)
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.Id, &t.Name, &t.Color, &t.UserId); err != nil {
			return model.Page[model.Tag]{}, fmt.Errorf("ListTags scan: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return model.Page[model.Tag]{}, fmt.Errorf("ListTags rows: %w", err)
	}
	if tags == nil {
		tags = []model.Tag{}
	}
	return model.Page[model.Tag]{
		Data:       tags,
		TotalCount: totalCount,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}, nil
}

func (r *PostgresStore) CreateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error) {
	if tag.Name == "" {
		return model.Tag{}, fmt.Errorf("CreateTag: name must not be empty: %w", model.ErrInvalidArgument)
	}
	if tag.Id == uuid.Nil {
		tag.Id = uuid.New()
	}

	const q = `
		INSERT INTO tags (id, name, color, user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, color, user_id`

	var created model.Tag
	err := r.db.QueryRow(ctx, q, tag.Id, tag.Name, tag.Color, scope.UserID()).
		Scan(&created.Id, &created.Name, &created.Color, &created.UserId)
	if err != nil {
		return model.Tag{}, fmt.Errorf("CreateTag: %w", err)
	}
	return created, nil
}

func (r *PostgresStore) UpdateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error) {
	if tag.Id == uuid.Nil {
		return model.Tag{}, fmt.Errorf("UpdateTag: id: %w", model.ErrInvalidArgument)
	}
	if tag.Name == "" {
		return model.Tag{}, fmt.Errorf("UpdateTag: name must not be empty: %w", model.ErrInvalidArgument)
	}

	const q = `
		UPDATE tags
		SET name = $2, color = $3
		WHERE id = $1 AND deleted_at IS NULL AND user_id IS NOT DISTINCT FROM $4
		RETURNING id, name, color, user_id`

	var updated model.Tag
	err := r.db.QueryRow(ctx, q, tag.Id, tag.Name, tag.Color, scope.UserID()).
		Scan(&updated.Id, &updated.Name, &updated.Color, &updated.UserId)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Tag{}, fmt.Errorf("UpdateTag %s: %w", tag.Id, model.ErrNotFound)
	}
	if err != nil {
		return model.Tag{}, fmt.Errorf("UpdateTag: %w", err)
	}
	return updated, nil
}

func (r *PostgresStore) DeleteTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("DeleteTag: id: %w", model.ErrInvalidArgument)
	}

	const q = `
		UPDATE tags
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL AND user_id IS NOT DISTINCT FROM $2`

	res, err := r.db.Exec(ctx, q, id, scope.UserID())
	if err != nil {
		return fmt.Errorf("DeleteTag: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("DeleteTag %s: %w", id, model.ErrNotFound)
	}
	return nil
}

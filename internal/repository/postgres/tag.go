package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/larssonoliver/inundated/internal/model"
)

// A task tag's archived state and color aren't stored on its row: it's
// archived while its task is closed, and it takes the color of the task's
// first regular tag by name, or model.DefaultTaskTagColor. These fragments
// read tags that way; they expect tags aliased t and the owning task (if
// any) LEFT JOINed as k, which tagFromSQL provides.
//
// Tag names sort case-insensitively, then by bytes, then by id, compared
// under the "C" collation rather than the database's own so the memory
// store can match the order. Case folding (lower, ILIKE) still follows the
// database's LC_CTYPE: under a UTF-8 locale it agrees with Go's
// strings.ToLower, but under the C locale only ASCII letters fold, so
// e.g. "Ärende" and "ärende" sort apart and search can't match them to
// each other there.
const (
	tagFromSQL = `tags t LEFT JOIN tasks k ON k.tag_id = t.id AND k.deleted_at IS NULL`

	tagArchivedAtSQL = `CASE WHEN k.id IS NULL THEN t.archived_at ELSE k.closed_at END`

	tagColorSQL = `CASE WHEN k.id IS NULL THEN t.color ELSE COALESCE((
			SELECT rt.color FROM task_tags kt
			JOIN tags rt ON rt.id = kt.tag_id AND rt.deleted_at IS NULL
			WHERE kt.task_id = k.id
			ORDER BY lower(rt.name) COLLATE "C", rt.name COLLATE "C", rt.id
			LIMIT 1
		), '` + model.DefaultTaskTagColor + `') END`

	tagColumnsSQL = `t.id, t.name, ` + tagColorSQL + `, t.user_id, ` + tagArchivedAtSQL + `, k.id`
)

func scanTag(row pgx.Row) (model.Tag, error) {
	var t model.Tag
	var archivedAt *time.Time
	if err := row.Scan(&t.Id, &t.Name, &t.Color, &t.UserId, &archivedAt, &t.TaskId); err != nil {
		return model.Tag{}, err
	}
	t.Archived = archivedAt != nil
	return t, nil
}

// tagListFilterSQL returns the WHERE conditions, each with a trailing
// "AND ", that ListTags adds for params, with any value they bind appended
// to args.
func tagListFilterSQL(params model.TagListParams, args []any) (string, []any) {
	var sql strings.Builder
	if !params.IncludeArchived {
		sql.WriteString(tagArchivedAtSQL + " IS NULL AND ")
	}
	switch params.Kind {
	case model.TagKindAll:
	case model.TagKindTask:
		sql.WriteString("k.id IS NOT NULL AND ")
	default:
		sql.WriteString("k.id IS NULL AND ")
	}
	if params.Query != "" {
		args = append(args, "%"+escapeLike(params.Query)+"%")
		fmt.Fprintf(&sql, `t.name ILIKE $%d ESCAPE '\' AND `, len(args))
	}
	if params.Ids != nil {
		args = append(args, params.Ids)
		fmt.Fprintf(&sql, `t.id = ANY($%d) AND `, len(args))
	}
	return sql.String(), args
}

// escapeLike escapes LIKE wildcards in s, using '\' as the escape character.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (r *PostgresStore) GetTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Tag, error) {
	if id == uuid.Nil {
		return model.Tag{}, fmt.Errorf("GetTag: id: %w", model.ErrInvalidArgument)
	}

	ownerSQL, args := ownerPredicate("t.user_id", scope, []any{id})
	q := `
		SELECT ` + tagColumnsSQL + `
		FROM ` + tagFromSQL + `
		WHERE t.id = $1 AND t.deleted_at IS NULL AND ` + ownerSQL

	t, err := scanTag(r.db.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Tag{}, fmt.Errorf("GetTag %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.Tag{}, fmt.Errorf("GetTag: %w", err)
	}
	return t, nil
}

func (r *PostgresStore) ListTags(ctx context.Context, scope model.OwnerScope, params model.TagListParams) (model.Page[model.Tag], error) {
	countFilterSQL, countArgs := tagListFilterSQL(params, nil)
	countOwnerSQL, countArgs := ownerPredicate("t.user_id", scope, countArgs)
	countQ := `
		SELECT COUNT(*)
		FROM ` + tagFromSQL + `
		WHERE t.deleted_at IS NULL AND ` + countFilterSQL + countOwnerSQL

	var totalCount int
	if err := r.db.QueryRow(ctx, countQ, countArgs...).Scan(&totalCount); err != nil {
		return model.Page[model.Tag]{}, fmt.Errorf("count tags: %w", err)
	}

	dataFilterSQL, args := tagListFilterSQL(params, []any{params.Limit, params.Offset})
	dataOwnerSQL, args := ownerPredicate("t.user_id", scope, args)
	q := `
		SELECT ` + tagColumnsSQL + `
		FROM ` + tagFromSQL + `
		WHERE t.deleted_at IS NULL AND ` + dataFilterSQL + dataOwnerSQL + `
		ORDER BY k.id IS NOT NULL, lower(t.name) COLLATE "C", t.name COLLATE "C", t.id
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return model.Page[model.Tag]{}, fmt.Errorf("ListTags: %w", err)
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		t, err := scanTag(rows)
		if err != nil {
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

	ownerSQL, args := ownerPredicate("user_id", scope, []any{tag.Id, tag.Name, tag.Color, tag.Archived})
	q := `
		UPDATE tags
		SET name = $2, color = $3,
			archived_at = CASE WHEN $4 THEN COALESCE(archived_at, now()) ELSE NULL END
		WHERE id = $1 AND ` + notTaskTagSQL + ` AND deleted_at IS NULL AND ` + ownerSQL + `
		RETURNING id, name, color, user_id, archived_at`

	var updated model.Tag
	var archivedAt *time.Time
	err := r.db.QueryRow(ctx, q, args...).
		Scan(&updated.Id, &updated.Name, &updated.Color, &updated.UserId, &archivedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Tag{}, fmt.Errorf("UpdateTag %s: %w", tag.Id, r.tagWriteMiss(ctx, scope, tag.Id))
	}
	if err != nil {
		return model.Tag{}, fmt.Errorf("UpdateTag: %w", err)
	}
	updated.Archived = archivedAt != nil
	return updated, nil
}

func (r *PostgresStore) DeleteTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("DeleteTag: id: %w", model.ErrInvalidArgument)
	}

	ownerSQL, args := ownerPredicate("user_id", scope, []any{id})
	q := `
		UPDATE tags
		SET deleted_at = now()
		WHERE id = $1 AND ` + notTaskTagSQL + ` AND deleted_at IS NULL AND ` + ownerSQL

	res, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("DeleteTag: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("DeleteTag %s: %w", id, r.tagWriteMiss(ctx, scope, id))
	}
	return nil
}

// notTaskTagSQL keeps UpdateTag and DeleteTag off task tags, which follow
// their task and can't be changed directly.
const notTaskTagSQL = `NOT EXISTS (SELECT 1 FROM tasks k WHERE k.tag_id = tags.id)`

// tagWriteMiss explains why UpdateTag or DeleteTag matched no row:
// model.ErrInvalidArgument if id is a task tag in scope, otherwise
// model.ErrNotFound.
func (r *PostgresStore) tagWriteMiss(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	ownerSQL, args := ownerPredicate("t.user_id", scope, []any{id})
	q := `
		SELECT EXISTS (
			SELECT 1 FROM tags t JOIN tasks k ON k.tag_id = t.id
			WHERE t.id = $1 AND t.deleted_at IS NULL AND ` + ownerSQL + `
		)`
	var taskTag bool
	if err := r.db.QueryRow(ctx, q, args...).Scan(&taskTag); err != nil {
		return fmt.Errorf("tagWriteMiss: %w", err)
	}
	if taskTag {
		return fmt.Errorf("tag belongs to a task: %w", model.ErrInvalidArgument)
	}
	return model.ErrNotFound
}

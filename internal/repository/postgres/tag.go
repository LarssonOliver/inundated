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

// An owned tag's archived state isn't stored on its row: it follows its
// owner (a task tag is archived while its task is closed), as the
// tag_owners view reports it. A derived tag's color is derived in the
// service layer (see model.DerivedTagColor), so the stored one is returned
// as is. These fragments read tags that way; they expect tags aliased t and
// its owner (if any) LEFT JOINed as o, which tagFromSQL provides.
//
// Tag names sort case-insensitively, then by bytes, then by id, compared
// under the "C" collation rather than the database's own so the memory
// store can match the order. Case folding (lower, ILIKE) still follows the
// database's LC_CTYPE: under a UTF-8 locale it agrees with Go's
// strings.ToLower, but under the C locale only ASCII letters fold, so
// e.g. "Ärende" and "ärende" sort apart and search can't match them to
// each other there.
const (
	tagFromSQL = `tags t LEFT JOIN tag_owners o ON o.tag_id = t.id`

	tagArchivedAtSQL = `CASE WHEN o.tag_id IS NULL THEN t.archived_at ELSE o.archived_at END`

	tagColumnsSQL = `t.id, t.name, t.color, t.user_id, ` + tagArchivedAtSQL + `, o.kind, o.owner_id`
)

func scanTag(row pgx.Row) (model.Tag, error) {
	var t model.Tag
	var archivedAt *time.Time
	var ownerKind *string
	var ownerId *uuid.UUID
	if err := row.Scan(&t.Id, &t.Name, &t.Color, &t.UserId, &archivedAt, &ownerKind, &ownerId); err != nil {
		return model.Tag{}, err
	}
	t.Archived = archivedAt != nil
	if ownerKind != nil && ownerId != nil {
		t.Owner = &model.TagOwner{Kind: model.TagOwnerKind(*ownerKind), Id: *ownerId}
	}
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
	if ownerKind, ok := params.Kind.OwnerKind(); ok {
		args = append(args, string(ownerKind))
		fmt.Fprintf(&sql, `o.kind = $%d AND `, len(args))
	} else if params.Kind != model.TagKindAll {
		sql.WriteString("o.tag_id IS NULL AND ")
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
		return model.Tag{}, fmt.Errorf("GetTag: %w", errNilId)
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
		ORDER BY o.tag_id IS NOT NULL, lower(t.name) COLLATE "C", t.name COLLATE "C", t.id
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

func (r *PostgresStore) ListDerivedTagSources(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) (map[uuid.UUID][]model.Tag, error) {
	out := map[uuid.UUID][]model.Tag{}
	if len(tagIds) == 0 {
		return out, nil
	}

	// A task tag's sources are its task's regular tags.
	ownerSQL, args := ownerPredicate("k.user_id", scope, []any{tagIds})
	query := `
		SELECT k.tag_id, t.id, t.name, t.color, t.user_id, t.archived_at
		FROM tasks k
		JOIN task_tags kt ON kt.task_id = k.id
		JOIN tags t ON t.id = kt.tag_id AND t.deleted_at IS NULL
		WHERE k.tag_id = ANY($1) AND k.deleted_at IS NULL AND ` + ownerSQL
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ListDerivedTagSources: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var derivedId uuid.UUID
		var source model.Tag
		var archivedAt *time.Time
		if err := rows.Scan(&derivedId, &source.Id, &source.Name, &source.Color, &source.UserId, &archivedAt); err != nil {
			return nil, fmt.Errorf("ListDerivedTagSources scan: %w", err)
		}
		source.Archived = archivedAt != nil
		out[derivedId] = append(out[derivedId], source)
	}
	return out, rows.Err()
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
		return model.Tag{}, fmt.Errorf("UpdateTag: %w", errNilId)
	}
	if tag.Name == "" {
		return model.Tag{}, fmt.Errorf("UpdateTag: name must not be empty: %w", model.ErrInvalidArgument)
	}

	ownerSQL, args := ownerPredicate("user_id", scope, []any{tag.Id, tag.Name, tag.Color, tag.Archived})
	q := `
		UPDATE tags
		SET name = $2, color = $3,
			archived_at = CASE WHEN $4 THEN COALESCE(archived_at, now()) ELSE NULL END
		WHERE id = $1 AND ` + notOwnedTagSQL + ` AND deleted_at IS NULL AND ` + ownerSQL + `
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
		return fmt.Errorf("DeleteTag: %w", errNilId)
	}

	ownerSQL, args := ownerPredicate("user_id", scope, []any{id})
	q := `
		UPDATE tags
		SET deleted_at = now()
		WHERE id = $1 AND ` + notOwnedTagSQL + ` AND deleted_at IS NULL AND ` + ownerSQL

	res, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("DeleteTag: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("DeleteTag %s: %w", id, r.tagWriteMiss(ctx, scope, id))
	}
	return nil
}

// notOwnedTagSQL keeps UpdateTag and DeleteTag off owned tags, which follow
// their owner and can't be changed directly.
const notOwnedTagSQL = `NOT EXISTS (SELECT 1 FROM tag_owners o WHERE o.tag_id = tags.id)`

// tagWriteMiss explains why UpdateTag or DeleteTag matched no row:
// model.ErrInvalidArgument if id is an owned tag in scope, otherwise
// model.ErrNotFound.
func (r *PostgresStore) tagWriteMiss(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	ownerSQL, args := ownerPredicate("t.user_id", scope, []any{id})
	q := `
		SELECT EXISTS (
			SELECT 1 FROM tags t JOIN tag_owners o ON o.tag_id = t.id
			WHERE t.id = $1 AND t.deleted_at IS NULL AND ` + ownerSQL + `
		)`
	var owned bool
	if err := r.db.QueryRow(ctx, q, args...).Scan(&owned); err != nil {
		return fmt.Errorf("tagWriteMiss: %w", err)
	}
	if owned {
		return fmt.Errorf("tag belongs to another item: %w", model.ErrInvalidArgument)
	}
	return model.ErrNotFound
}

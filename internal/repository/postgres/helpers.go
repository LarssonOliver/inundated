package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/larssonoliver/inundated/internal/model"
)

// errNilId is what by-id lookups return for the zero UUID. No row ever has
// that id, so it's a miss like any other, not a malformed request; returning
// early just saves the round trip.
var errNilId = fmt.Errorf("zero id: %w", model.ErrNotFound)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func ownerPredicate(column string, scope model.OwnerScope, args []any) (sql string, newArgs []any) {
	if id := scope.UserID(); id != nil {
		return fmt.Sprintf("%s = $%d", column, len(args)+1), append(args, *id)
	}
	return column + " IS NULL", args
}

// archivedFilter returns the SQL fragment (with a trailing "AND ") that
// excludes archived rows, or an empty string when includeArchived is true.
// This is baked into the query text rather than bound as a parameter: an
// `archived_at IS NULL OR $n` predicate can't use an index on archived_at,
// and folding it into a query-shaped string avoids hand-syncing its
// placeholder number against the rest of the positional args.
func archivedFilter(includeArchived bool) string {
	if includeArchived {
		return ""
	}
	return "archived_at IS NULL AND "
}

// noAssociatedTags is a tagsInScope alreadyAssociated getter for a brand new
// project/timespan, which by definition has no pre-existing associations.
func noAssociatedTags() ([]uuid.UUID, error) {
	return nil, nil
}

// sqlConditionBuilder accumulates "AND <fragment> $N" clauses alongside
// their bound values, numbering each placeholder from args already added
// (optionally seeded via newSQLConditionBuilder) - so a caller adding a
// condition never has to compute $N by hand or keep it in sync with a
// sibling query built the same way.
//
// ownerPredicate isn't built on this type (it's used throughout this
// package with its own append-and-return-args convention), so it must
// still be called last, passed this builder's own Args(), to keep its
// placeholder numbering correct.
type sqlConditionBuilder struct {
	sql  strings.Builder
	args []any
}

// newSQLConditionBuilder seeds the builder with args already bound ahead of
// any condition it will add (e.g. LIMIT/OFFSET), so its own placeholder
// numbers continue that sequence instead of starting over at $1.
func newSQLConditionBuilder(seedArgs []any) *sqlConditionBuilder {
	return &sqlConditionBuilder{args: seedArgs}
}

// add appends " AND <fragment> $N" to the builder's SQL, bound to value.
// fragment is everything before the placeholder, e.g. "end_time >".
func (b *sqlConditionBuilder) add(fragment string, value any) {
	b.args = append(b.args, value)
	fmt.Fprintf(&b.sql, " AND %s $%d", fragment, len(b.args))
}

// addExpr appends " AND <expr>" to the builder's SQL, bound to value, with
// every "$?" in expr replaced by value's placeholder. Use it when the
// placeholder can't go last, e.g. inside a subquery.
func (b *sqlConditionBuilder) addExpr(expr string, value any) {
	b.args = append(b.args, value)
	b.sql.WriteString(" AND " + strings.ReplaceAll(expr, "$?", fmt.Sprintf("$%d", len(b.args))))
}

func (b *sqlConditionBuilder) SQL() string { return b.sql.String() }
func (b *sqlConditionBuilder) Args() []any { return b.args }

// linkedTagIds returns the live tag IDs linked to each of ids in a tag join
// table (e.g. project_tags, keyed by fkCol project_id), keyed by id. Links to
// deleted tags are kept (tag deletion is soft) but never read back, so a
// read-modify-write never sends one back. Callers pass r.db for a standalone
// read, or the transaction's Querier to read within it (e.g. alongside a
// concurrent tagsInScope check).
func linkedTagIds(ctx context.Context, q Querier, table, fkCol string, ids []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	out := make(map[uuid.UUID][]uuid.UUID, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	query := `
		SELECT l.` + fkCol + `, l.tag_id
		FROM ` + table + ` l
		JOIN tags t ON t.id = l.tag_id AND t.deleted_at IS NULL
		WHERE l.` + fkCol + ` = ANY($1)
		ORDER BY l.` + fkCol + `, l.tag_id`
	rows, err := q.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("%s tag ids: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, tagId uuid.UUID
		if err := rows.Scan(&id, &tagId); err != nil {
			return nil, fmt.Errorf("%s tag ids scan: %w", table, err)
		}
		out[id] = append(out[id], tagId)
	}
	return out, rows.Err()
}

// setLinkedTags replaces the tags linked to id in a tag join table (see
// linkedTagIds) with tagIds, inserting them in one statement.
func setLinkedTags(ctx context.Context, q Querier, table, fkCol string, id uuid.UUID, tagIds []uuid.UUID) error {
	if _, err := q.Exec(ctx, `DELETE FROM `+table+` WHERE `+fkCol+` = $1`, id); err != nil {
		return fmt.Errorf("%s delete: %w", table, err)
	}
	if len(tagIds) == 0 {
		return nil
	}
	if _, err := q.Exec(ctx,
		`INSERT INTO `+table+` (`+fkCol+`, tag_id) SELECT $1, unnest($2::uuid[])`,
		id, tagIds,
	); err != nil {
		return fmt.Errorf("%s insert: %w: %w", table, model.ErrInvalidReference, err)
	}
	return nil
}
